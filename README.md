# Go Application Template
This is a simple template for developing Go applications. It includes a Makefile for building, testing, and analyzing the code, as well as a `.gitignore` file to exclude unnecessary files from version control.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
<br>

## Todos after cloning

### Files
#### Naming
Use the `init.py` script to set toolname, module name, author, and year.
```bash
python init.py
```

Alternatively, you may manually update the following files:
- [ ] Update the `module` in `go.mod` to your project's module path.
- [ ] Update `tool-name` in `Makefile` to your desired tool name.
- [ ] Update `<tool-name>` in `main.go`, `cmd/version/command.go` to your desired tool name.
- [ ] Update the copyright year and name in `LICENSE`.

#### Content
- [ ] Update this `README.md` to reflect your project details.
- [ ] When **NOT** using Docker:
  - [ ] Delete the `Dockerfile`
  - [ ] Remove the `docker` dependabot config at `.github/dependabot.yml`
  - [ ] Remove the steps starting at `- name: Set up QEMU` in `.github/workflows/release.yml`
- [ ] When using Docker:
  - [ ] Update `<tool-name>` in `Dockerfile` and `.github/workflows/release.yml` (if not already done by `init.py`)

### GitHub
- [ ] Import the `ruleset.json` as a branch protection rule in your GitHub repository settings.
  - [ ] Delete the `ruleset.json` file after importing.
- [ ] Enable `Allow auto-merge` in the general settings of your GitHub repository.
- [ ] Enable `Automatically delete head branches` in the general settings of your GitHub repository.
- [ ] Create the `Dependencies` tag

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details
