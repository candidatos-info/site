import sys
import re

"""This script gets PROJECT_ID, EMAIL, PASSWORD, SITE_URL and SECRET environment variables used on app engine from Github Secrets and
replace on app.yaml."""

app_engine_file = "app.yaml"

if __name__ == "__main__":
    if len(sys.argv) != 8:
        sys.exit("invalid number of arguments: {}".format(len(sys.argv)))
    project_id = sys.argv[1]
    email = sys.argv[2]
    password = sys.argv[3]
    site_url = sys.argv[4]
    secret = sys.argv[5]
    election_year = sys.argv[6]
    update_profile = sys.argv[7]
    file_content = ""
    with open(app_engine_file, "r") as file:
        app_engine_file_content = file.read()
        line = re.sub(r"##PROJECT_ID", project_id, app_engine_file_content)
        line = re.sub(r"##EMAIL", email, line)
        line = re.sub(r"##PASSWORD", password, line)
        line = re.sub(r"##SITE_URL", site_url, line)
        line = re.sub(r"##SECRET", secret, line)
        line = re.sub(r"##ELECTION_YEAR", election_year, line)
        line = re.sub(r"##UPDATE_PROFILE", update_profile, line)
        file_content = line
    with open(app_engine_file, "w") as file:
        file.write(file_content)
