# Flexo

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-blue)

<img align="right" src="./branding/Flexo.png" width="150" alt="Flexo logo">

Music Discord bot written in [Go](https://go.dev) using [Disgo](https://github.com/disgoorg/disgo) and [Lavalink](https://github.com/lavalink-devs/Lavalink).

> [!WARNING]
> In development. May contain bugs.

## Features

- Music playback from YouTube, Spotify, etc.
- Hybrid commands (slash + prefix)
- Poise-inspired framework

## Commands

| Command | Description   | Usage                            |
| :------ | :------------ | :------------------------------- |
| `play`  | Play music    | `/play <song>` or `!play <song>` |
| `skip`  | Skip track    | `/skip` or `!skip`               |
| `ping`  | Check latency | `/ping` or `!ping`               |

## Requirements

- Go 1.25+
- Discord bot token
- Lavalink server with: [LavaSrc](https://github.com/topi314/LavaSrc), [LavaQueue](https://github.com/topi314/LavaQueue), [YouTube Source](https://github.com/lavalink-devs/youtube-source)

## Installation

```bash
git clone https://github.com/goland-express/Flexo
cd flexo
cp .env.example .env
# Add Discord token + Lavalink config to .env
go run .
```

---

## Tech Stack

| Library                                               | Purpose                |
| :---------------------------------------------------- | :--------------------- |
| [Disgo](https://github.com/disgoorg/disgo)            | Discord API wrapper    |
| [DisgoLink](https://github.com/disgoorg/disgolink)    | Lavalink client        |
| [Lavalink](https://github.com/lavalink-devs/Lavalink) | Audio streaming server |

## TODO

- [ ] Queue management commands
- [ ] Volume control
- [ ] Loop/repeat modes
- [ ] Playlist support
- [ ] Web dashboard

## Contributing

Pull requests are welcome. For major changes, open an issue first to discuss what you want to change.
