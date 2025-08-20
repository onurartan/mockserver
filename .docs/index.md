#### Run Project -> Shell
```bash
go run . start --config example/mockserver.json
```


#### Build Project -> Shell
```bash
# STEP 1
# It will create a binary file according to your operating system.
go build

# STEP 2
# for WINDOWS
# You can use the built file in this way
mockserver.exe start --config example/mockserver.json


# for LINUX/MAC
# You can use the built file in this way
./mockserver start --config example/mockserver.json
```


#### Build Npm Package -> Shell

```bash
# STEP 1
# open git bash
# A platform-specific build file is created 
# npm/bin/...
./build.sh


# STEP 2
# To test the project globally, add it to the global package with pnpm.
pnpm link --global


# STEP 3
# You can now test mockserver.
# Specify the file path after --config.
mockserver start --config example/mockserver.json
```