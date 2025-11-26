# Quick Start - Install magi

> [!TIP]
> magi is installable via [instl.sh](https://instl.sh).\
> You just have to run the following command and you're ready to go!

<!-- tabs:start -->

#### ** Windows **

### Windows Command

```powershell
iwr instl.sh/MagdielCAS/magi-cli/windows | iex
```

#### ** Linux **

### Linux Command

```bash
curl -sSL instl.sh/MagdielCAS/magi-cli/linux | bash
```

#### ** macOS **

### macOS Command

```bash
curl -sSL instl.sh/MagdielCAS/magi-cli/macos | bash
```

#### ** Compile from source **

### Compile from source with Golang

?> **NOTICE**
To compile magi from source, you have to have [Go](https://golang.org/) installed.

Compiling magi from source has the benefit that the build command is the same on every platform.\
It is not recommended to install Go only for the installation of magi.

```command
go install github.com/MagdielCAS/magi-cli@latest
```

<!-- tabs:end -->
