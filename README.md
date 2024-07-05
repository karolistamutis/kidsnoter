# kidsnoter

kidsnoter is a tool for downloading and organizing albums from Kidsnote, a popular childcare communication platform.

## 👶 Motivation

My children attended a daycare which uses kidsnote.com to communicate with the parents. The daycare posts a ton of albums
with loads of images but I saw no way to take them out - that's no good. Losing years of photo and video memories just
because your children change daycare (or grow up) is a scary thought. Therefore I wrote this tool that can be run once
or continuously to download all albums and save their descriptions for future smiles.

## Features

- Automatically downloads highest fidelity photos and videos from Kidsnote.
- - Smartly skips files that do not require replacing.
- Organizes photos and videos into a structured directory hierarchy:
- - $CHILDNAME/$YEAR/$MONTH_$ALBUMID_$ALBUMNAME
- Generates Markdown files with album information so that full title and description is preserved.
- Docker support for easy cross platform deployment.
- - But does not require Docker to run.
- Different verbosity logging levels.
- Instrumented with Prometheus for dashboards and alerting.
- Good API citizen with limited concurrency, exponential backoff and random jitter.

## 🚀 I AM NOT TECHNICAL, JUST TELL ME HOW TO RUN THIS

## Installation Guide for Windows Users

If you're not familiar with technical setups, follow these steps to get kidsnoter running on your Windows computer:

### 1. Install Docker Desktop

1. Download Docker Desktop for Windows from the official website:
   https://www.docker.com/products/docker-desktop
2. Double-click the installer file you downloaded (probably named "Docker Desktop Installer.exe").
3. Follow the installation wizard:
    - If asked, ensure the "Use WSL 2 instead of Hyper-V" option is selected.
    - Keep all other settings at their default values.
4. Click "Ok" to start the installation.
5. Once installation is complete, click "Close" to finish the installer.
6. Docker Desktop will start automatically. You may see a popup asking for permissions; click "Ok" to allow Docker to make changes.

### 2. Create Necessary Folders and Files

1. 📸 Create a folder for your albums:
   - Right-click on your Desktop
   - Select "New" > "Folder"
   - Name the folder "albums"
2. ⚙️ Create a configuration file:
   - Right-click on your Desktop
   - Select "New" > "Text Document"
   - Name the file "config.yaml" (make sure to change the extension from .txt to .yaml)
3. Edit the config file:
   - Right-click on "config.yaml" and select "Open with" > "Notepad"
   - Copy and paste the following into the file:
     ```yaml
     username: your_kidsnote_username
     password: your_kidsnote_password
     album_dir: /data/kidsnote
     ```
   - Replace `your_kidsnote_username` and `your_kidsnote_password` with your actual kidsnote.com login details.
   - Save the file and close Notepad.

### 3. Run kidsnoter

1. Open "Command Prompt" (you can search for it in the Start menu)
2. Copy and paste the following command, then press Enter:

```shell
docker run -v %USERPROFILE%\Desktop\config.yaml:/config/config.yaml -v %USERPROFILE%\Desktop\albums:/data/kidsnote ghcr.io/karolistamutis/kidsnoter:latest download-albums -vvv
```

3. Wait for the command to finish, depending on amount of photos and your Internet connection speed it may take up to a few hours or more. You should see content appearing in the `albums` folder you've created on your Desktop earlier as it runs.

## 🖥️ Installation for Nerds (no Docker)

1. Ensure you have Go 1.22 or later installed:

```shell
go version
```

2. Clone the repository:

```shell
git clone https://github.com/karolistamutis/kidsnoter.git
cd kidsnoter
```

3. Build the binary:

```shell
make build
```

4. Create a `config.yaml` file in the project root:

```yaml
username: your_kidsnote_username
password: your_kidsnote_password
album_dir: /path/to/save/albums
sync_interval: 12h
```

5. Check that it works by listing your children

```shell
./kidsnoter list-children
```

6. Run once or keep the process to sync periodically according to `sync_interval`

Once:
```shell
./kidsnoter download-albums
```
Continuous:
```shell
./kidsnoter serve
```

The commands above are equivalent, I've implemented `serve` specifically for running within Docker because Cronjobs or Systemd timers are not a good fit there. Feel free to run just `download-albums` with Cron or Systemd if you so wish!

If you will run within Docker, expose port `:9091`, you'll be able to scrape `/metrics` Prometheus endpoint.

### ⚙️ Environment variables

The following environment variables can be used to override `config.yaml` settings. You can also skip having `config.yaml` with the below set:
* `KIDSNOTE_USERNAME`
* `KIDSNOTE_PASSWORD`
* `KIDSNOTE_ALBUM_DIR`
* `KIDSNOTE_SYNC_INTERVAL` in [Go's Time.Duration format](https://pkg.go.dev/time#ParseDuration)

### Commands

* `list-children` will list children under your account.
* `list-albums` will list all albums for all children.
* * given `--child-id` or `--child-name` will limit to a single child.
* `download-albums` will download all albums for all children.
* * given `--child-id` or `--child-name` will limit to a single child.
* `serve` will download all albums for all children and repeat the process according to `sync_interval` config parameter.

### Commandline flags

* `-v` or `-vv` or `-vvv` or `-vvvv` for most verbose log output
* `--overwrite` bool flag for `sync` or `download-albums` to always download files and rewrite album descriptions.

Contributing
Contributions are welcome! Please feel free to submit a Pull Request.

## License
This project is licensed under the MIT License.

## Acknowledgements
Thanks to the kidsnote.com team for their platform.

This project is not officially affiliated with or endorsed by kidsnote.com.