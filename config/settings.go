// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package config

import (
    "strings"

    "github.com/Sirupsen/logrus"
    "github.com/arsham/expvastic/lib"
    "github.com/arsham/expvastic/reader/expvar"
    "github.com/arsham/expvastic/reader/self"
    "github.com/arsham/expvastic/recorder/elasticsearch"
    "github.com/spf13/viper"
)

// routeMap looks like this:
// {
//     route1: {readers: [my_app, self], recorders: [elastic1]}
//     route2: {readers: [my_app], recorders: [elastic1, file1]}
// }
type routeMap map[string]route
type route struct {
    readers   []string
    recorders []string
}

// ConfMap holds the relation between readers and recorders.
type ConfMap struct {
    // Readers contains a map of reader names to their configuration.
    Readers map[string]ReaderConf

    // Recorders contains a map of recorder names to their configuration.
    Recorders map[string]RecorderConf

    // Routes contains a map of reader names to a list of recorders.
    // map["red1"][]string{"rec1", "rec2"}: means whatever is read from red1, will be shipped to both rec1 and rec2
    Routes map[string][]string
}

func checkSettingsSect(log *logrus.Logger, v *viper.Viper) error {
    if v.IsSet("settings.debug_evel") {
        newLevel, ok := v.Get("settings.debug_evel").(string)
        if !ok {
            return &StructureErr{"debug_level", "should be a string", nil}
        }
        *log = *lib.GetLogger(newLevel)
    }
    return nil
}

// LoadYAML loads the settings from the configuration file
func LoadYAML(log *logrus.Logger, v *viper.Viper) (*ConfMap, error) {
    var (
        readerKeys   map[string]string
        recorderKeys map[string]string
        routes       routeMap
        err          error
    )
    if len(v.AllSettings()) == 0 {
        return nil, EmptyConfigErr
    }
    if v.IsSet("settings") {
        err = checkSettingsSect(log, v)
        if err != nil {
            return nil, &StructureErr{"settings", "", err}
        }
    }

    if readerKeys, err = getReaders(v); err != nil {
        return nil, err
    }
    if recorderKeys, err = getRecorders(v); err != nil {
        return nil, err
    }

    if routes, err = getRoutes(v); err != nil {
        return nil, err
    }

    if err = checkAgainstReadRecorders(routes, readerKeys, recorderKeys); err != nil {
        return nil, err
    }

    return loadConfiguration(v, routes, readerKeys, recorderKeys)
}

//readers is a map of keyName:typeName
func getReaders(v *viper.Viper) (readers map[string]string, err error) {
    if !v.IsSet("readers") {
        return nil, newNotSpecifiedErr("readers", "", nil)
    }
    readers = make(map[string]string)
    for reader, settings := range v.GetStringMap("readers") {
        _ = settings
        switch rType := v.GetString("readers." + reader + ".type"); rType {
        case "self":
            readers[reader] = rType
        case "expvar":
            readers[reader] = rType
        case "":
            fallthrough
        default:
            return nil, newNotSpecifiedErr(reader, "type not defined", nil)
        }

    }
    return
}

//recorders is a map of keyName:typeName
func getRecorders(v *viper.Viper) (recorders map[string]string, err error) {
    if !v.IsSet("recorders") {
        return nil, newNotSpecifiedErr("recorders", "", nil)
    }
    recorders = make(map[string]string)

    for recorder, settings := range v.GetStringMap("recorders") {
        _ = settings
        switch rType := v.GetString("recorders." + recorder + ".type"); rType {
        case "elasticsearch":
            recorders[recorder] = rType
        case "":
            fallthrough
        default:
            return nil, newNotSpecifiedErr(recorder, "type not defined", nil)
        }
    }
    return
}

func getRoutes(v *viper.Viper) (routes routeMap, err error) {
    if !v.IsSet("routes") {
        return nil, newNotSpecifiedErr("routes", "", nil)
    }
    routes = make(map[string]route)
    for name := range v.GetStringMap("routes") {
        rot := route{}
        for recRedType, list := range v.GetStringMapStringSlice("routes." + name) {
            for _, target := range list {
                if strings.Contains(target, ",") {
                    return nil, newRoutersErr(recRedType, "not an array or single value", nil)
                }
                if recRedType == "readers" {
                    rot.readers = append(rot.readers, target)
                } else if recRedType == "recorders" {
                    rot.recorders = append(rot.recorders, target)
                }
            }
            routes[name] = rot
        }
        if len(routes[name].readers) == 0 {
            return nil, newRoutersErr("readers", "is empty", nil)
        }
        if len(routes[name].recorders) == 0 {
            return nil, newRoutersErr("recorders", "is empty", nil)
        }
    }
    return
}

func checkAgainstReadRecorders(routes routeMap, readerKeys, recorderKeys map[string]string) error {
    for _, section := range routes {
        for _, reader := range section.readers {
            if !lib.StringInMapKeys(reader, readerKeys) {
                return newRoutersErr("routers", reader+" not in readers", nil)
            }
        }
        for _, recorder := range section.recorders {
            if !lib.StringInMapKeys(recorder, recorderKeys) {
                return newRoutersErr("routers", recorder+" not in recorders", nil)
            }
        }
    }
    return nil
}

func loadConfiguration(v *viper.Viper, routes routeMap, readerKeys, recorderKeys map[string]string) (*ConfMap, error) {
    confMap := &ConfMap{
        Readers:   make(map[string]ReaderConf, len(readerKeys)),
        Recorders: make(map[string]RecorderConf, len(recorderKeys)),
    }

    for name, reader := range readerKeys {
        r, err := parseReader(v, reader, name)
        if err != nil {
            return nil, err
        }
        confMap.Readers[name] = r
    }

    for name, recorder := range recorderKeys {
        r, err := readRecorders(v, recorder, name)
        if err != nil {
            return nil, err
        }
        confMap.Recorders[name] = r
    }
    confMap.Routes = mapReadersRecorders(routes)
    return confMap, nil
}

func parseReader(v *viper.Viper, readerType, name string) (ReaderConf, error) {
    switch readerType {
    case "expvar":
        return expvar.FromViper(v, name, "readers."+name)
    case "self":
        return self.FromViper(v, name, "readers."+name)
    }
    return nil, notSupportedErr(readerType)
}

func readRecorders(v *viper.Viper, recorderType, name string) (RecorderConf, error) {
    switch recorderType {
    case "elasticsearch":
        return elasticsearch.FromViper(v, name, "recorders."+name)
    }
    return nil, notSupportedErr(recorderType)
}

// For future reference, something like this might happen:

// readers:
//     app_0:
//     app_1:
//     app_2:
//
// recorders:
//     elastic_0:
//     elastic_1:
//     elastic_2:
//     elastic_3:
//
// routes:
//     route1:
//         readers:
//             - app_0
//             - app_2
//         recorders:
//             - elastic_1
//     route2:
//         readers:
//             - app_0
//         recorders:
//             - elastic_1
//             - elastic_2
//             - elastic_3
//     route2:
//         readers:
//             - app_1
//             - app_2
//         recorders:
//             - elastic_1
//             - elastic_0
//
//
// We need to turn it into:
//
// app_0: [elastic_1, elastic_2, elastic_3]
// app_1: [elastic_1, elastic_0]
// app_2: [elastic_1, elastic_0]
func mapReadersRecorders(routes routeMap) map[string][]string {
    // We don't know how this matrix will be, let's go dynamic!
    // This looks ugly. The whole logic should change. But it doesn't have any impact on the program, it just runs once.
    readerMap := make(map[string][]string)
    for _, route := range routes {
        // Add the readers to the map
        for _, redName := range route.readers {
            // now iterate through the recorders and add them
            for _, recName := range route.recorders {
                if _, ok := readerMap[redName]; !ok {
                    readerMap[redName] = []string{recName}
                } else {
                    readerMap[redName] = append(readerMap[redName], recName)
                    // Shall we go another level deep??? :p
                    // I'm kidding, seriously, refactor this thing
                    // Do you know why the chicken crossed the road?
                    // There was a few nested eggs on the other side!
                    // ok, back to the business.
                    // BTW ask me why I left these comments.
                }
            }
        }
    }
    // Let's cleanup
    resultMap := make(map[string][]string)
    for reader, recorders := range readerMap {
        checkMap := make(map[string]bool)
        for _, recorder := range recorders {
            if _, ok := checkMap[recorder]; !ok {
                checkMap[recorder] = true
                if _, ok := resultMap[reader]; !ok {
                    resultMap[reader] = []string{recorder}
                } else {
                    resultMap[reader] = append(resultMap[reader], recorder)
                    // Remember that chicker? It's roasted now.
                }
            }
        }
    }
    return resultMap
}
