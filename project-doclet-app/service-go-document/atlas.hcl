data "external_schema" "gorm" {
  program = ["go", "run", "./cmd/atlas"]
}

env "local" {
  src = data.external_schema.gorm.url
  dev = "postgres://doclet:doclet@localhost:5432/doclet_dev?sslmode=disable"
  url = "postgres://doclet:doclet@localhost:5432/doclet?sslmode=disable"
  migration {
    dir = "file://migrations"
  }
}
