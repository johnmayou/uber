variable "REGISTRY" { default = "ghcr.io/johnmayou/uber" }
variable "TAG" { }

target "base" {
  dockerfile = "Dockerfile"
  context    = "../"
}

target "auth" {
  inherits = ["base"]
  args     = { SERVICE = "auth" }
  tags     = ["${REGISTRY}/services/auth:${TAG}"]
}

target "trip" {
  inherits = ["base"]
  args     = { SERVICE = "trip" }
  tags     = ["${REGISTRY}/services/trip:${TAG}"]
}

target "user" {
  inherits = ["base"]
  args     = { SERVICE = "user" }
  tags     = ["${REGISTRY}/services/user:${TAG}"]
}
