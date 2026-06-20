<!--
         ▄▀▀▀▄ ▄▀▀▀▄    █▄   ▄█  ▄▀▄  ▀▀▀▀█ ▄▀▀▀▄ ▄▀▀▀▄ ▄▀▀▀▄
         █     █   █    █ ▀▄▀ █ █   █    ▄▀ █   █ █     █
         █     █   █ ▄▄ █     █ █   █  ▄▀   █   █ █      ▀▀▀▄
         █ ▀▀█ █   █    █     █ █▄▄▄█ █     █   █ █ ▀▀█     █
         ▀▄▄▄▀ ▀▄▄▄▀    █     █ █   █ █▄▄▄▄ ▀▄▄▄▀ ▀▄▄▄▀ ▀▄▄▄▀

                          ▄▄ ▄       ▄▄▀    ▀▄▄
                       ▄▄▄█▀ █       █▄▄▄██▄▄▄█
                       █ ██▀▀█          ████
                        █▀▀▀█         ▄▀▀██▀▀▄
                      █▀▀   ▀▀        ▀▀    ▀▀
-->
# Go-Mazogs

A Go port of **Mazogs**, the classic maze game originally written by Don Priestley
for the Sinclair ZX-81. This implementation is based on the [disassembly by
Paul Farrow](http://www.fruitcake.plus.com/Sinclair/ZX81/Disassemblies/Mazogs.htm).

*DISCLAIMER: This is an unofficial, independent rewrite of the 1982 game Mazogs
using Go and SDL3. It is not affiliated with, endorsed by, or associated with
the original creators or copyright holders of Mazogs.*


## About the game

Mazogs is a real-time maze exploration game where the player must navigate a
randomly generated maze to find a treasure and return to the exit. Along the
way, the player encounters mazogs (hostile creatures), prisoners who may reveal
the route to the treasure, and swords that can be used in combat.

The game features three difficulty levels and a limited number of moves,
requiring the player to balance exploration with efficiency.


## Controls

| Key        | Action                   |
|------------|--------------------------|
| A / H / ←  | Move left                |
| D / J / →  | Move right               |
| W / ↑      | Move up                  |
| X / S / ↓  | Move down                |
| V          | Display mini-map view    |
| Y          | Request situation report |


## Building

Requires Go 1.25+ and SDL3 development libraries.

```sh
go build
```


## Running

```sh
./mazogs
```

To use Wayland directly (instead of XWayland):

```sh
SDL_VIDEO_DRIVER=wayland ./mazogs
```
