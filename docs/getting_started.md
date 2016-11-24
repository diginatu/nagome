Getting started
===============

Installation
------------

Assume you have the go developing environment.

~~~ sh
go get github.com/diginatu/nagome
~~~

To make normal plugin
---------------------

Let's make a new TCP plugin.

Make your plugin template with running the command.

~~~ sh
nagome --makeplug awesome_plugin
~~~

Now you can find plugin.yml generated in your plugin directory.
Modify it like this.

~~~ yml
name: awesome_plugin
description: very awsome nagome plugin
version: "1.0"
author: Me
method: tcp
exec:
- '{{path}}/awesome_plugin'
- '{{port}}'
- '{{no}}'
nagomever: ""
subscribe:
- nagome
~~~


