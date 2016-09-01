Nagome
======

[![Circle CI](https://circleci.com/gh/diginatu/nagome.svg?style=svg)](https://circleci.com/gh/diginatu/nagome)

Advanced NicoLive Comment Viewer written in go

Nagome はアドバンストでクロスプラットフォームなニコニコ生放送用 コメントビューア（ニコ生 コメビュ）です。

[Viqo](https://github.com/diginatu/Viqo) よりさらに自由なコメビュを目指しています。

UIを分離し、UIのインターフェースを提供するのが特徴です。


Licence
-------

MIT Licence


Dependencies
------------

 + gopkg.in/yaml.v2 : Apache Licence 2.0
 + diginatu/nagome/nicolive
   - gopkg.in/xmlpath.v2 : LGPLv3
   - gopkg.in/yaml.v2 : Apache Licence 2.0


Tasks
-----

 - [ ] add broadcast info to open message
 - [ ] translation function
 - [ ] document (plugin)
 - [ ] function for NicoLive
    - [ ] getting user name and storing
    - [ ] alert connection
    - [ ] get new Waku

