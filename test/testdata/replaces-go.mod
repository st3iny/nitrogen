module test-module.com/module

go 1.26

replace my-dep1 => ../../my-dep1
replace my-dep2 => ../../my-dep2

require (
    my-dep1 v1.0.0
    my-dep2 v1.0.0
)
