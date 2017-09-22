// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package config

import (
	"strings"

	"github.com/arsham/expvastic/internal"
	"github.com/arsham/expvastic/reader/expvar"
	"github.com/arsham/expvastic/reader/self"
	"github.com/arsham/expvastic/recorder/elasticsearch"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	selfReader            = "self"
	expvarReader          = "expvar"
	elasticsearchRecorder = "elasticsearch"
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

	// Routes contains a map of recorder names to a list of readers.
	// map["rec1"][]string{"red1", "red2"}: means whatever is read
	// from red1 and red2, will be shipped to rec1.
	Routes map[string][]string
}

// Checks the application scope settings. Applies them if defined.
// If the log level is defined, it will replace a new logger with the provided one.
func checkSettingsSect(log *internal.Logger, v *viper.Viper) error {
	if v.IsSet("settings.log_level") {
		newLevel, ok := v.Get("settings.log_level").(string)
		if !ok {
			return &StructureErr{"log_level", "should be a string", nil}
		}
		*log = *internal.GetLogger(newLevel)
	}
	return nil
}

// LoadYAML loads the settings from the configuration file.
// It returns any errors returned from readers/recorders. Please
// refer to their documentations.
func LoadYAML(log *internal.Logger, v *viper.Viper) (*ConfMap, error) {
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
		return nil, errors.WithMessage(err, "readerKeys")
	}
	if recorderKeys, err = getRecorders(v); err != nil {
		return nil, errors.WithMessage(err, "recorderKeys")
	}

	if routes, err = getRoutes(v); err != nil {
		return nil, errors.WithMessage(err, "routes")
	}

	if err = checkAgainstReadRecorders(routes, readerKeys, recorderKeys); err != nil {
		return nil, errors.WithMessage(err, "checkAgainstReadRecorders")
	}

	return loadConfiguration(v, log, routes, readerKeys, recorderKeys)
}

// readers is a map of keyName:typeName
// typeName is not the recorder's type, it's the extension name, e.g. expvar.
func getReaders(v *viper.Viper) (map[string]string, error) {
	readers := make(map[string]string)

	if !v.IsSet("readers") {
		return nil, newNotSpecifiedErr("readers", "", nil)
	}

	for reader := range v.GetStringMap("readers") {
		switch rType := v.GetString("readers." + reader + ".type"); rType {
		case selfReader:
			readers[reader] = rType
		case expvarReader:
			readers[reader] = rType
		case "":
			fallthrough
		default:
			return nil, newNotSpecifiedErr(reader, "type", nil)
		}
	}
	return readers, nil
}

// recorders is a map of keyName:typeName
// typeName is not the recorder's type, it's the extension name, e.g. elasticsearch.
func getRecorders(v *viper.Viper) (map[string]string, error) {
	recorders := make(map[string]string)

	if !v.IsSet("recorders") {
		return nil, newNotSpecifiedErr("recorders", "", nil)
	}

	for recorder := range v.GetStringMap("recorders") {
		switch rType := v.GetString("recorders." + recorder + ".type"); rType {
		case elasticsearchRecorder:
			recorders[recorder] = rType
		case "":
			fallthrough
		default:
			return nil, newNotSpecifiedErr(recorder, "type", nil)
		}
	}
	return recorders, nil
}

func getRoutes(v *viper.Viper) (routeMap, error) {
	routes := make(map[string]route)
	if !v.IsSet("routes") {
		return nil, newNotSpecifiedErr("routes", "", nil)
	}

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
	return routes, nil
}

// Checks all apps in routes are mentioned in the readerKeys and recorderKeys.
func checkAgainstReadRecorders(routes routeMap, readerKeys, recorderKeys map[string]string) error {
	for _, section := range routes {
		for _, reader := range section.readers {
			if !internal.StringInMapKeys(reader, readerKeys) {
				return newRoutersErr("routers", reader+" not in readers", nil)
			}
		}

		for _, recorder := range section.recorders {
			if !internal.StringInMapKeys(recorder, recorderKeys) {
				return newRoutersErr("routers", recorder+" not in recorders", nil)
			}
		}
	}

	return nil
}

func loadConfiguration(v *viper.Viper, log internal.FieldLogger, routes routeMap, readerKeys, recorderKeys map[string]string) (*ConfMap, error) {
	confMap := &ConfMap{
		Readers:   make(map[string]ReaderConf, len(readerKeys)),
		Recorders: make(map[string]RecorderConf, len(recorderKeys)),
	}

	for name, reader := range readerKeys {
		r, err := parseReader(v, log, reader, name)
		if err != nil {
			return nil, errors.Wrap(err, "reader keys")
		}
		confMap.Readers[name] = r
	}

	for name, recorder := range recorderKeys {
		r, err := readRecorders(v, log, recorder, name)
		if err != nil {
			return nil, errors.Wrap(err, "recorder keys")
		}
		confMap.Recorders[name] = r
	}

	confMap.Routes = mapReadersRecorders(routes)
	return confMap, nil
}

func parseReader(v *viper.Viper, log internal.FieldLogger, readerType, name string) (ReaderConf, error) {
	switch readerType {
	case expvarReader:
		rc, err := expvar.FromViper(v, log, name, "readers."+name)
		return rc, errors.Wrap(err, "parsing reader")
	case selfReader:
		rc, err := self.FromViper(v, log, name, "readers."+name)
		return rc, errors.Wrap(err, "parsing reader")
	}
	return nil, notSupportedErr(readerType)
}

func readRecorders(v *viper.Viper, log internal.FieldLogger, recorderType, name string) (RecorderConf, error) {
	switch recorderType {
	case elasticsearchRecorder:
		rc, err := elasticsearch.FromViper(v, log, name, "recorders."+name)
		return rc, errors.Wrap(err, "read-recorders loading from viper")
	}
	return nil, notSupportedErr(recorderType)
}

func mapReadersRecorders(routes routeMap) map[string][]string {
	// We don't know how this matrix will be, let's go dynamic!
	// This looks ugly. The whole logic should change. But it doesn't have any impact on the program, it just runs once.
	recorderMap := make(map[string][]string)
	for _, route := range routes {
		// Add the recorders to the map
		for _, recName := range route.recorders {
			// now iterate through the readers and add them
			for _, readName := range route.readers {
				if _, ok := recorderMap[recName]; !ok {
					recorderMap[recName] = []string{readName}
				} else {
					recorderMap[recName] = append(recorderMap[recName], readName)
					// Shall we go another level deep??? :p
					// I'm kidding, seriously, refactor this thing
					// Do you know why the chicken crossed the road?
					// There was a few nested eggs on the other side!
					// Okay, back to the business.
					// BTW ask me why I left these comments.
				}
			}
		}
	}

	// Let's clean up
	resultMap := make(map[string][]string)
	for recName, reds := range recorderMap {
		checkMap := make(map[string]bool)
		for _, readName := range reds {
			if _, ok := checkMap[readName]; !ok {
				checkMap[readName] = true
				if _, ok := resultMap[recName]; !ok {
					resultMap[recName] = []string{readName}
				} else {
					resultMap[recName] = append(resultMap[recName], readName)
					// Remember that chicken? It's roasted now.
				}
			}
		}
	}
	return resultMap
}
