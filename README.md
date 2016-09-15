Nagome
======

[![Circle CI](https://circleci.com/gh/diginatu/nagome.svg?style=svg)](https://circleci.com/gh/diginatu/nagome)

Advanced NicoLive Comment Viewer written in go.

Nagome has no UI but API to communicate with plugins.
So it doesn't depend on platforms and still its look.
You can make various UI like native desktop app on any platform, modern app on the browser, even as Vim plugin.
It can be also used for daemon like a bot.

Installation
------------

Assume you have the go developing environment.

~~~ sh
go get https://github.com/diginatu/nagome
~~~

Licence
-------

[MIT Licence](LICENSE)

Dependencies
------------

+   gopkg.in/yaml.v2 : Apache Licence 2.0
+   diginatu/nagome/nicolive
    -   gopkg.in/xmlpath.v2 : LGPLv3
    -   gopkg.in/yaml.v2 : Apache Licence 2.0

Tasks
-----

+   [ ] message that clear comments
+   [ ] nagomever
+   [ ] add broadcast info to the message open
+   [ ] translation function
+   [ ] document (plugin)
+   [ ] function for NicoLive
    -   [ ] getting user name and storing
    -   [ ] alert connection
    -   [ ] get new Waku

