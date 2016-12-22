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

Let's make a new stdio plugin.

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
method: std
exec:
- '{{path}}/awesome_plugin'
nagomever: ""
subscribe:
- nagome
~~~

Then put your executable into plugin directory and name it `awesome_plugin`.
And you can already send and receive [Nagome Messages](nagome_message.md).

Also, more information about plugins can be found [here](plugin.md).
