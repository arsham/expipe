# Changelog

## v1.0-rc1
## Release Candidate 1
- Removes backoff values.
- Tested nil values in MapConvert (fixes #28)
- Separate a part of the Main function to able to test it
- Generates an up-to-date version of kibana config file.

## v0.11.0
### Reshaping the Engine
- Changed the behaviour of the Engine to have one reader to multiple recorders. (fixes #32)
- Changed the Engine to be an interface.
- Introduced the Operator type.
- Fixed loading unused readers and recorders from config files.
- Added examples for engine.
- Sending jobs are now by value.

## v0.10.0
### Mostly refactoring
- Cleaned up and made more precise, better and concurrent safe parallel tests.
- Added testing with testdata to config.
- Changed some internal names.
- Changed some packages into better locations.

### v0.9.2
### All about tests
- Removed ginkgo from readers and recorders.
- Added more tests and clean up old ones.
- Added .codecov.yml configuration file.
- Various fixes on datatype
- Fixed some documentation problems.
- Changed Config structs to be used only for reading from config files.
- Removed stale code.

### v0.9.1
# Changed elasticsearch, expvar, self and testings config to use Config arguments.
# Removed coveralls script from travis
# Changed the coverage badge

### v0.9.0

- Changed dependencies and refactored code.
- Replaced flag package with go-flags.
- Split the main function so we can safely remove the helpers.
- Moved most of main.go functionalities to app package.
- Changed the container and datatype apis to be writers.
- Removed the Err field from Container.
- Removed bdd bits from where they should test units.

### v0.8.4
- Moved config package as a part of internal.
- Moved datattype interfaces into their own file.
- Used global TimeStampFormat variable to format times.
- Addd tests to datatypes.
- Movde map_reader internal tests into another file.
- Checking for logger should not be changed during construction of readers and recorders.
- Removed unused expvar value.
- Exported ErrInvalidURL and removed its unused method receiver.
- Added ginkgo to tests.
- Refactored recorders and readers to use ginkgo.

### v0.8.1
- Removed context from reader/recorder.
- Set up readers, recorders and the engine with config options.
- Defer tests back to reader/recorder tests.
- Renamed the package from expvastic to expipe.
- Renamed lib package to internal.
- Made tests "race detection" friendly.
- Moved tests related to logrus into an exempted file due to internal race conditions.
- Moved token and datatype packages to internal package.
- Added vendor to .gitignore file.
- Added bootstrap tests.
