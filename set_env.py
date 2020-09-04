import sys
import re

"""This script get PROJECT_ID environment variable used on app engine from Github Secrets and
replace on app.yaml."""

app_engine_file = "app.yaml"

if __name__ == "__main__":
    if len(sys.argv) != 2:
        sys.exit("invalid number of arguments: {}".format(len(sys.argv)))
    project_id = sys.argv[1]
    file_content = ""
    with open(app_engine_file, "r") as file:
        app_engine_file_content = file.read()
        line = re.sub(r"##PROJECT_ID", project_id, app_engine_file_content)
        file_content = line
    with open(app_engine_file, "w") as file:
        file.write(file_content)
