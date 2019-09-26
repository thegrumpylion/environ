## Structured environment parsing for Go

Marshal structured data to and unmarshal structured data from environment variables without resulting to serialization formats like json, yaml, etc.

### Unmarshal from string slice

```go
env := []string{
    "ADDR=localhost",
    "PORT=8080",
}

conf := &struct {
    Addr string
    Port int
}{}

environ.UnmarshalEnv(conf, env)

fmt.Printf("Host: %s:%d\n", conf.Addr, conf.Port)
```

Prints out

```shell
Host: localhost:8080
```

### Unmarshal from os environment

```go
conf := &struct {
    Addr string
    Port int
}{}

environ.UnmarshalEnviron(conf)

fmt.Printf("Host: %s:%d\n", conf.Addr, conf.Port)
```

```shell
$ ADDR=localhost PORT=8099 ./example
Host: localhost:8099
```

`environ.UnmarshalEnviron(var)` is equivalent of calling `environ.UnmarshalEnv(var, os.Environ())`

### Namespacing with prefixes

Environment variables can be prefixed

```go
env := []string{
    "_PFX_ADDR=localhost",
    "_PFX_PORT=8088",
}

conf := &struct {
    Addr string
    Port int
}{}

environ.UnmarshalEnvPfx(conf, env, "_PFX_")

fmt.Printf("Host: %s:%d\n", conf.Addr, conf.Port)
```

Prints out

```shell
Host: localhost:8088
```

### Unsetting variables after parsing

```go
os.Clearenv()
os.Setenv("ADDR", "localhost")
os.Setenv("PORT", "8080")

conf := &struct {
    Addr string
    Port int
}{}

fmt.Println("Before", os.Environ())

environ.UnmarshalEnvironAndUnset(conf)

fmt.Println("After", os.Environ())
```

Prints out

```shell
Before [ADDR=localhost PORT=8080]
After []
```