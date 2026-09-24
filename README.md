# marauderctl

<div align="center">
<a href="https://github.com/9g10f/marauderctl"><img src="https://img.shields.io/github/stars/9g10f/marauderctl?style=flat&logo=github&color=orange" alt="Stars"></a>
<a href="https://github.com/9g10f/marauderctl/network/members"><img src="https://img.shields.io/github/forks/9g10f/marauderctl?style=flat&logo=github&color=orange" alt="Forks"></a>
<a href="https://codeberg.org/9g10f/marauderctl"><img src="https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fcodeberg.org%2Fapi%2Fv1%2Frepos%2F9g10f%2Fmarauderctl&query=%24.stars_count&label=stars&logo=codeberg" alt="Stars"></a>
<a href="https://codeberg.org/9g10f/marauderctl"><img src="https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fcodeberg.org%2Fapi%2Fv1%2Frepos%2F9g10f%2Fmarauderctl&query=%24.forks_count&label=forks&logo=codeberg" alt="Forks"></a>
<a href="https://github.com/9g10f/marauderctl/tags"><img src="https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fcodeberg.org%2Fapi%2Fv1%2Frepos%2F9g10f%2Fmarauderctl%2Ftags&query=%24.0.name&label=tag&color=white" alt="Tag"></a>
<a href="https://github.com/9g10f/marauderctl?tab=GPL-3.0-1-ov-file"><img src="https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fapi.github.com%2Frepos%2F9g10f%2Fmarauderctl%2Flicense&query=%24.license.spdx_id&label=license&color=white" alt="License"></a>
<br>
<img src="https://img.shields.io/badge/platform-Linux%20%7C%20Windows%20%7C%20BSD%20%7C%20macOS-green" alt="Platforms">
</div>

<br>

<div align="center">
<a href="https://codeberg.org/9g10f/marauderctl"> <img src="https://img.shields.io/badge/Codeberg%20mirror-black?logo=codeberg&style=for-the-badge" alt="Codeberg mirror"> </a>
</div>

---

marauderctl is the CLI backend for Marauder Launcher. You can use it to install & run regular or pirated games through a CLI. marauderctl relies on custom scripts to install games with our main server (storage for the scripts) being [marauder.k.vu](https://marauder.k.vu).

> [!IMPORTANT]
> This project is still under heavy development and has not even reached v0.1, use it with caution.

## Installation
You can install `marauderctl` through Go if you have it installed:
```
go install https://codeberg.org/9g10f/marauderctl@latest
```

## Usage
You can use `marauderctl` to install games through it's `game-id` and it's `game-version` on your server, but it's best to first always check the scripts for malicious code:
```
marauderctl query <game-id>@[game-version/latest] script
```
This prints the script of the game you selected. After checking the script install it by running:
```
marauderctl install <game-id>@[game-version/latest]
```
You can also run your games with:
```
marauderctl start <game-id>@[game-version/latest]
```

What this looks like in practice:
> ```
> $ marauderctl query 0 script
> $ marauderctl install 0
> $ marauderctl start 0
> ```

## License
marauderctl is free and licensed under the [GNU General Public License v3.0](LICENSE).