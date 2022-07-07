variable "GORELEASER_XX_BASE" {
  default = "crazymax/goreleaser-xx:edge"
}

target "_commons" {
  args = {
    GORELEASER_XX_BASE = GORELEASER_XX_BASE
  }
}

group "default" {
  targets = ["image-local"]
}

target "base" {
  inherits = ["_commons"]
  target = "base"
  output = ["type=cacheonly"]
}

target "image" {
  inherits = ["_commons"]
  tags = ["ledfx_go:local"]
}

target "image-local" {
  inherits = ["image"]
  output = ["type=docker"]
}

target "image-all" {
  inherits = ["image"]
  platforms = [
    "linux/amd64",
    "linux/arm/v6",
    "linux/arm/v7",
    "linux/arm64"
  ]
}

target "artifact" {
  inherits = ["_commons"]
  target = "artifact"
  output = ["./dist"]
}

target "artifact-all" {
  inherits = ["artifact"]
  platforms = [
    "linux/amd64",
    "linux/arm/v6",
    "linux/arm/v7",
    "linux/arm64",
  ]
}