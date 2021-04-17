cmg - Chaotic Media Gopher
==========================

cmg is a gopher frontend for media.ccc.de

Caveat
------

For historical reasons most (if not all) gopher clients expect the server to
speak ASCII. Nonetheless cmg speaks UTF-8. Terminal based clients (like lynx(1)
or gopher(1)) running in an UTF-8 environment are okay with this. Your mileage
may vary, though.

Streaming
---------

Gopher clients in general know nothing about media streaming. The easiest way to
stream a recording via cmg is to use curl(1) and a pipe to your preferred media
player:

    curl '<gopher-URI>' | mplayer -

