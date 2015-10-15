[![Build Status](https://travis-ci.org/shingetsu-gou/shingetsu-gou.svg?branch=master)](https://travis-ci.org/shingetsu-gou/shingetsu-gou)
[![GoDoc](https://godoc.org/github.com/shingetsu-gou/shingetsu-gou?status.svg)](https://godoc.org/github.com/shingetsu-gou/shingetsu-gou)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/shingetsu-gou/shingetsu-gou/master/LICENSE)


# Gou(合) 

## Overview

Gou（[合](https://ja.wikipedia.org/wiki/%E5%90%88_%28%E5%A4%A9%E6%96%87%29)) is a clone of P2P anonymous BBS shinGETsu saku in golang.

The word "Gou(合)" means [conjunction](https://en.wikipedia.org/wiki/Astrological_aspect) in Japanese, when an aspect is an angle the planets make to each other in the horoscope.

Yeah, the sun and moon are in conjunction during the new moon(新月, 朔）.


## License

MIT License

Original Program comes from [saku](https://github.com/shingetsu/saku), which is under [2-clause BSD license](https://github.com/shingetsu/saku/blob/master/LICENSE)
Copyrighted by 2005-2015 shinGETsu Project.

See also

 * www/bootstrap/css/bootstrap.min.css
 * www/jquery/MIT-LICENSE.txt
 * www/jquery/jquery.min.js
 * www/jquery/jquery.lazy.min.js
 * www/jquery/spoiler/authors.txt


## Requirements

* git
* go 1.4+

are required to compile.

## Installation

    $ mkdir gou
    $ cd gou
    $ mkdir src
    $ mkdir bin
    $ mkdir pkg
    $ exoprt GOPATH=`pwd`
    $ go get github.com/shingetsu-gou/shingetsu-gou
	
Or you can download executable binaries from [here](https://github.com/shingetsu-gou/shingetsu-gou/releases).

# Differences from Original Saku

1. mch(2ch interface) listens to the same port as admin.cgi/gateway.cgi/serve.cgi/thread.cgi. dat_port setting in config is ignored.
2. For now Gou doesn't consider synchronism. Caches in disk may be broken. This problem would be fixed if everyone thinks Gou is useful :) .
3. Gou can try to open port by uPnP and NAT-PMP. You can disable this function by setting [Gateway] enable_nat:false in saku.ini, which is true by default.
4. URL for 2ch interface /2ch_***/subject.txt in saku is /2ch/***/subject.txt in Gou.


# Note

Files 

* in template/ directory
* in www/ directory
* in file/ directory

are embeded into the exexutable binary in https://github.com/shingetsu-gou/shingetsu-gou/releases.
If these files are not found on your disk, Gou automatically expands these to the disk.
Once expanded, you can change these files as you wish.

This is for easy-use of Gou; just get a binary, and run it!

# Contribution

Improvements to the codebase and pull requests are encouraged.


