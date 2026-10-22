# ilias-sync-webdav
Download your ILIAS course files using WebDAV.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
<br>

## Setup

### Binary
The easiest way to get started is by downloading the [latest release](https://github.com/matthiasharzer/ilias-sync-webdav/releases/latest) for your platform.

## Usage
To start syncing one or multiple courses, you need to set up a config-file and run the following command:

```bash
ilias-sync-webdav sync -c ilias.cfg
````

The config file should look like this:

```ini
; Define some global settings
[DEFAULT]
url = https://ilias.example.com/webdav.php/...    ; The base URL of your ILIAS WebDAV endpoint
directory = /path/to/download/directory           ; The local directory where files will be downloaded (grouped by course)
file_already_exists_strategy = <skip | overwrite> ; (optional) What to do if a file already exists in the download directory. Default is 'skip'

; Authentication methods (choose one):
[auth:basic]
username = your_username  ; Your ILIAS username
password = your_password  ; Your ILIAS password

[auth:file]
path = /path/to/credentials.txt  ; Path to a file containing your ILIAS username in the first line and your password in the second line

[auth:env] ; Use ILIAS_USERNAME and ILIAS_PASSWORD environment variables for authentication

; Define the courses you want to sync
[sync:<course1-name>] 
id = <course1-id>   ; The ID of the course you want to sync

[sync:<course2-name>]
id = <course2-id>   ; The ID of the course you want to sync
```

> The course ID can be found in the URL of the course when you are logged in to ILIAS. It is usually a number at the end of the URL like `&ref_id=<course-id>`.

Exactly one authentication method must be configured.

### Verified ILIAS URLs
The following ILIAS URLs have been verified to work with this syncer:
- https://ilias.h-ka.de/webdav.php/HSKA

## ILIAS WebDAV Limitations
ILIAS WebDAV disallows some characters in file names, such as `:`, `*`, `?`, `"`, `<`, `>`, `|`, and `/`. Files containing these characters are not exposed via WebDAV and will not be downloaded. Though the syncer will log a warning for each file that is skipped, it is recommended to check the course files in the ILIAS web interface to ensure that all files have been downloaded.

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details
