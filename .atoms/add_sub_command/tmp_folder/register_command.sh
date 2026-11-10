LINE=$(find-line find "// ! Register commands" cmd/app.go)
find-line insert "tmp_folder/1_0_app.go" "cmd/app.go" $LINE