[![GoDoc](https://godoc.org/github.com/udhos/equalfile?status.svg)](http://godoc.org/github.com/udhos/equalfile)
[![Go Report Card](https://goreportcard.com/badge/github.com/udhos/equalfile)](https://goreportcard.com/report/github.com/udhos/equalfile)
[![Travis Build Status](https://travis-ci.org/udhos/equalfile.svg?branch=master)](https://travis-ci.org/udhos/equalfile)
[![Circle CI](https://circleci.com/gh/udhos/equalfile.svg?style=shield&circle-token=:circle-token)](https://circleci.com/gh/udhos/equalfile)
[![gocover](http://gocover.io/_badge/github.com/udhos/equalfile)](http://gocover.io/github.com/udhos/equalfile)

About Equalfile 
===============

Equalfile is a pure Go package for comparing files.

Install
=======

## Recipe with Modules (since Go 1.11)

Clone outside of GOPATH:

    git clone https://github.com/udhos/equalfile
    cd equalfile

Run tests:

    go test

Install example application 'equal':

    go install ./equal

Run example application:

    ~/go/bin/equal

## Recipe without Modules (before Go 1.11)

Set up GOPATH as usual:

    export GOPATH=$HOME/go
    mkdir $GOPATH

Get equalfile package:

    go get github.com/udhos/equalfile

Install example application 'equal':

    go install github.com/udhos/equalfile/equal

Run example application:

    $GOPATH/bin/equal

Usage
=====

Add the import path:

    import "github.com/udhos/equalfile"

See: [equalfile GoDoc API](https://godoc.org/github.com/udhos/equalfile)

Example Application
===================

See example application: [equal source code](https://github.com/udhos/equalfile/blob/master/equal/main.go)
