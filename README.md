# ue4pak
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FVilsol%2Fue4pak.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2FVilsol%2Fue4pak?ref=badge_shield)


```
ue4pak parses and extracts data from UE4 Pak files

Usage:
  ue4pak [command]

Available Commands:
  class-tree  Read paks and output their class trees
  extract     Extract provided asset paths
  help        Help about any command
  test        Test parse the provided paks

Flags:
      --colors       Force output with colors
  -h, --help         help for ue4pak
      --log string   The log level to output (default "info")
      --no-preload   Do not preload data (slower, but guaranteed to read)
  -p, --pak string   The path to pak file (supports glob) (required)

Use "ue4pak [command] --help" for more information about a command.
```

## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FVilsol%2Fue4pak.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2FVilsol%2Fue4pak?ref=badge_large)