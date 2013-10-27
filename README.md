Summer
======

A Go framework for performing dependency injection with the goal of reducing object
creation boilerplate.


## Installation

    go get github.com/felixalias/summer

To run included tests:

    cd $GOROOT/src/pkg/github.com/felixalias/summer
    go test


## Status

Summer is functional for my needs. All features listed are fully implemented, working, and testing.
It is capable of:

  - Injecting dependencies into your structs by name
  - Injecting dependencies automatically by matching types

In addition,

  - Summer is [fully documented](http://godoc.org/github.com/felixalias/summer)
  - Summer has thorough test coverage


## Quick Start

See the examples for [PerformInjections](http://godoc.org/github.com/felixalias/summer/#Container.PerformInjections) and [InjectInto](http://godoc.org/github.com/felixalias/summer/#Container.InjectInto).

```go
import "github.com/felixalias/summer"

type MyService struct {
    MyDependency     string `summer:"StringDependency"`
    MyAutoDependency string `summer:",auto"`
}
service := new(MyService)

container := NewContainer()
container.Add("injected value", "StringDependency")
container.Add(service, "")

ok := container.PerformInjections()
```


## Documentation

Documentation can be found at [godoc.org/github.com/felixalias/summer](http://godoc.org/github.com/felixalias/summer).


## License

Copyright © 2013 Felix Jodoin

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
