import sys
import re

"""This script gets EMAIL, PASSWORD, SITE_URL, SECRET, ELECTION_YEAR, UPDATE_PROFILE, DB_NAME and DB_URL environment variables used on app engine from Github Secrets and
replace on app.yaml."""

app_engine_file = "app.yaml"

if __name__ == "__main__":
    if len(sys.argv) != 9:
        sys.exit("invalid number of arguments: {}".format(len(sys.argv)))
    email = sys.argv[1]
    password = sys.argv[2]
    site_url = sys.argv[3]
    secret = sys.argv[4]
    election_year = sys.argv[5]
    update_profile = sys.argv[6]
    db_name = sys.argv[7]
    db_url = sys.argv[8]
    file_content = ""
    with open(app_engine_file, "r") as file:
        app_engine_file_content = file.read()
        line = re.sub(r"##EMAIL", email, app_engine_file_content)
        line = re.sub(r"##PASSWORD", password, line)
        line = re.sub(r"##SITE_URL", site_url, line)
        line = re.sub(r"##SECRET", secret, line)
        line = re.sub(r"##ELECTION_YEAR", election_year, line)
        line = re.sub(r"##UPDATE_PROFILE", update_profile, line)
        line = re.sub(r"##DB_NAME", db_name, line)
        line = re.sub(r"##DB_URL", db_url, line)
        file_content = line
    with open(app_engine_file, "w") as file:
        file.write(file_content)
