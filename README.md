# Radix

An implementation of a radix tree in Go. See

> Donald R. Morrison. "PATRICIA -- practical algorithm to retrieve              
> information coded in alphanumeric". Journal of the ACM, 15(4):514-534,        
> October 1968    

Or the [wikipedia article](http://en.wikipedia.org/wiki/Radix_tree).

## Usage

Get the package:

	$ go get github.com/miekg/radix

Import the package:

	import (
		"github.com/miekg/radix"
	)

You can use the tree as a key-value structure, where every node's can have its
own value (as shown in the example below), or you can of course just use it to
look up strings, like so:

	r := radix.New()
	r.Insert("foo", true)
        x, e := r.Find("foo")
        if e {
	    fmt.Printf("foo is contained: %v\n", x.Value)
        }

### Documentation

For full package documentation, visit http://go.pkgdoc.org/github.com/miekg/radix.

## License

This code is licensed under a BSD License:

    Copyright (c) 2012 Alexander Willing and Miek Gieben. All rights reserved.
    Use of this source code is governed by a BSD-style
    license that can be found in the LICENSE file.


