# Autobindings 

Autobindings is a simple extention to the amazing library [Binding](https://github.com/mholt/binding). So binding is a reflectionless data binding for Go's net/http. For that developer has to write a FieldMap function which is used by this library to map the incoming JSON from the request to the struct fields.

## What autobinding does ?
So it automatically creates FieldMap function for your struct. 

## How to use ?
Just add this line to all of your files which has struct and for which you want to create a FieldMap function

```
//go:generate autobindings <file_name>
```

Or

From command line just run

```
autobindings <file_name>
```

## How to install ?
```
go install github.com/rainingclouds/autobindings
```

## How does it happens ?
Using the power of go generate :)
It creates <struct_name>_bindings.go for every struct in the given file. This file contains FieldMap function for that struct. This function is used by [Binding](https://github.com/mholt/binding) library to perform the mapping.

## Okay so what is missing ?
* It doesn't support Embedded fields yet.

## Happy to help
akshay@rainingclouds.com
[@akshay_deo](https://twitter.com/akshay_deo)
