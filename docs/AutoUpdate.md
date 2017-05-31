# Auto Update

Slick has update functionality built in. If enabled, slick will do a few things on startup:

- Make a request to the github api for the latest release: `https://api.github.com/repos/1egoman/slick/releases/latest`
- If the latest version dosn't match the current version:
  - Download the latest version into memory, for the given `GOOS` and `GOARCH`.
  - Discover where slime is currently executing from with `os.Executable()`
  - Overwrite the current executable with the downloaded one.
  - Prompt to restart slick.

## Enabling
Since the feature is a little buggy still and can cause problems in development, it's currently
disabled by default. Enable by adding `Set("AutoUpdate", "true")` to your
[`.slickrc`](Scripting.md). Feel free to [open issues](https://github.com/1egoman/issues/new) pertaining to this feature.
