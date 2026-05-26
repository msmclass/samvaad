# Mageutil

A collection of common helpers meant for use in [magefiles](https://magefile.org/).

## Examples:

Single command

```go
func Build() error {
    ctx := context.Background()
    return mageutil.RunDir(ctx, "cmd/server", "go build -o ../../bin/samvaad-server")
}
```

Multiple commands

```go
const gstVersion = 1.20.3

func BuildDocker(version string) error {
    return mageutil.Run(context.Background(),
        fmt.Sprintf("docker pull samvaad/gstreamer:%s-dev", gstVersion),
        fmt.Sprintf("docker pull samvaad/gstreamer:%s-prod", gstVersion),
        fmt.Sprintf("docker build -t samvaad/egress:v%s -f build/Dockerfile .", version),
    )
}
```

Updating repos

```go
func BuildSamvaad() error {
    ctx := context.Background()
	
    dir, err := filepath.Abs("..")
    if err != nil {
        return err
    }

    if err = mageutil.CloneRepo("samvaad", "samvaad", dir); err != nil {
        return err
    }
    
    dir, err = filepath.Abs("../samvaad")
    if err != nil {
        return err
    }
	
    return mageutil.RunDir(ctx, dir, "mage build")
}
```

Tools

```go
func Generate() error {
    ctx := context.Background()
    err := mageutil.InstallTool("github.com/google/wire/cmd/wire", "latest", false)
    if err != nil {
        return err
    }
    return mageutil.Run(ctx, "go generate ./...")
}
```

Group

```go
func RunSamvaadWithEgress() error {
    ctx := context.Background()
    group := mageutil.NewGroup(ctx)
    group.Go(func () error {
        return RunDir(ctx, "../samvaad", "bin/samvaad-server --dev")
    })
    group.Go(func () error {
        return Run(ctx, "docker run --rm -e EGRESS_CONFIG_FILE=/out/local.yaml -v ~/samvaad/egress/test:/out samvaad/egress")
    })
    group.Wait()
}
```
