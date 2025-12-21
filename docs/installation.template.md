# Quick Start - Install {{ .Name }}

> [!TIP]
> {{ .Name }} is installable via our installation script or using Go.

<!-- tabs:start -->

#### ** Windows **

### Windows Command

```powershell
go install github.com/MagdielCAS/magi-cli@latest
```

#### ** Linux **

### Linux Command

```bash
curl -sSL https://raw.githubusercontent.com/MagdielCAS/magi-cli/main/scripts/install.sh | bash
```

#### ** macOS **

### macOS Command

```bash
curl -sSL https://raw.githubusercontent.com/MagdielCAS/magi-cli/main/scripts/install.sh | bash
```

#### ** Compile from source **

### Compile from source with Golang

?> **NOTICE**
To compile {{ .Name }} from source, you have to have [Go](https://golang.org/) installed.

Compiling {{ .Name }} from source has the benefit that the build command is the same on every platform.\
It is not recommended to install Go only for the installation of {{ .Name }}.

```command
go install github.com/{{ .ProjectPath }}@latest
```

<!-- tabs:end -->
