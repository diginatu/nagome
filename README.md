Nagome
======

[![GoDoc](https://godoc.org/github.com/diginatu/nagome?status.svg)](https://godoc.org/github.com/diginatu/nagome)
[![Build Status](https://travis-ci.org/diginatu/nagome.svg?branch=master)](https://travis-ci.org/diginatu/nagome)

Advanced NicoLive Comment Viewer written in go.

Nagome has no UI but API to communicate with plugins.
So it doesn't depend on platforms or environments.
You can make various UIs like native desktop app on any platform, modern app on the browser, even as Vim plugin.
It can be also used for daemon like bots.

Install
------- 

Assume you have the go developing environment.

~~~ sh
go get -u github.com/diginatu/nagome
~~~

Document
--------

[Index](docs/README.md)

Nagome is initial development yet.
APIs may be changed.
But some features work now.

Licence
-------

[MIT License](LICENSE)

Dependencies
------------

+   gopkg.in/yaml.v2 : Apache Licence 2.0
+   diginatu/nagome/nicolive
    -   gopkg.in/xmlpath.v2 : LGPLv3
    -   gopkg.in/yaml.v2 : Apache Licence 2.0
    -   github.com/syndtr/goleveldb : 2-Clause BSD License

Contribution
------------

Contribution is welcome, about anything like fixing issues, adding new features, etc.

You can contact me via [my twitter](https://twitter.com/diginatu).
Also, [my niconico community](http://com.nicovideo.jp/community/co2345471) here.

Feel free to send a message and tell me what feature you want to work in.
Or, about plugins.

Tasks
-----

* [ ] Add a function to automatically avoid busy TCP port
* [ ] Check the settings value nagomever for a plugin
* [ ] Add a Nagome Message for quitting Nagome itself
* [ ] Add a feature to Add/remove a plugin dynamically
* [ ] Make Direct Domain available to plugin
* [ ] Translation of the UI
* [ ] Add more document
