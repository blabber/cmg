cmg - Chaotic Media Gopher
==========================

cmg is a gopher frontend for media.ccc.de

Caveat
------

For historical reason most (if not all) gopher clients expect the server
to speak ASCII. Nonetheless cmg speaks UTF-8. Terminal based clients
(like lynx(1) or gopher(1)) running in an UTF-8 environment are okay
with this. Your mileage may vary, though.

Streaming
---------

Gopher clients in general know nothing about media streaming. The
easiest way to stream a recording via cmg is to use curl(1) and a pipe
to your preferred media player:

    curl '<gopher-URI>' | mplayer -

Live installation
-----------------

* gopher://gopher.raumzeitlabor.org/
* [via HTTP-Proxy (no dedicated gopher client needed)](http://gopher.floodgap.com/gopher/gw?a=gopher%3A%2F%2Fgopher.raumzeitlabor.org)

As bandwidth is limited on gopher.raumzeitlabor.org, the links to the
recording are leaving gopherspace and point via HTTP to the media.ccc.de
CDN. For the real gopher feel, consider running your own instance with
recordings provided as item type '9'.
