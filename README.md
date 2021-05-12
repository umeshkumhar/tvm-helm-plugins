# tvm-helm-plugins

`tvm-upgrade` plugin is the helm plugin which is used to run the pre-upgrade job to allow the upgrade of 
`TrilioVaultManager` from version v2.0.x to version v2.1.x

## Install

```
helm plugin install <URL which I update later>
```

## Running a pre-upgrade job

```
helm tvm-upgrade --release <name> --namespace <ns>
```

