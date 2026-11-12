LINE=$(atom-util find "// ! Register commands" cmd/app.go)
atom-util insert -b "tmp_folder/1_0_app.go" "cmd/app.go" $LINE