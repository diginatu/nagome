Nagome
======

[![GoDoc](https://godoc.org/github.com/diginatu/nagome?status.svg)](https://godoc.org/github.com/diginatu/nagome)
[![Build Status](https://travis-ci.org/diginatu/nagome.svg?branch=master)](https://travis-ci.org/diginatu/nagome)
[![codecov](https://codecov.io/gh/diginatu/nagome/branch/master/graph/badge.svg)](https://codecov.io/gh/diginatu/nagome)

General comment viewer for live streams written in go.

Nagome doesn't have user interface.
It only provides APIs to communicate with plugins.
A UI application can be implemented as a plugin using this API.
Because it doesn't depend on environment, you can make various kind of UI (e.g. native desktop app, web app, CUI).
You can also easily create an app something like a bot which interact with comments.

Features
--------

### Supported Live Stream

* YouTube
* ~~NicoNico Live~~ (outdated)

UI Implementation
-----------------

### [Nagome Electron](https://github.com/diginatu/nagome-electron)

Desktop app implementation using Nagome Web UI below.
All dependencies is packed as an app.  You can just download it and use it.

### [Nagome WebUI](https://github.com/diginatu/nagome-webui)

Static web SPA.
It can be used as a part of an app or embedded in another web UI.
It runs with nagome-webui-server which run as a Nagome plugin and provides web socket.

Plugins
-------

Nagome plugins recomended to add github topic "nagome-plugin".

https://github.com/topics/nagome-plugin

Install
-------

Assume you have the go developing environment.

~~~ sh
go install github.com/diginatu/nagome
~~~

Document
--------

[Index](docs/README.md)

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

You can contact me via [my twitter](https://twitter.com/diginatu).

Feel free to send a message and tell me what feature you want to work in or even your plugins you want to make.

Tasks
-----

* [ ] Show error when a plugin failed to load
* [ ] Check "nagomever" in the plugin setting
* [ ] Add a feature to Add/remove a plugin dynamically
* [ ] Localization
* API
    * [ ] Make Direct Domain available to plugin
    * [ ] Add a Nagome Message for quitting Nagome itself

Release
-------

``` sh
git tag [version] -a
git push --tags
```
