import re
import subprocess
import datetime
import os


def str_prompt(prompt: str, default_value: str = "") -> str:
	if default_value:
		return input(f"{prompt} [{default_value}]: ") or default_value
	else:
		return input(prompt + ":")

def replace_template(files: list[str], template_value: str, new_value: str):
	for file in files:
		try:
			with open(file, "r", encoding="utf-8") as f:
				content = f.read()
		except FileNotFoundError:
			print(f"Skipping missing file: {file}")
			continue
		content = content.replace(template_value, new_value)
		with open(file, "w", encoding="utf-8") as f:
			f.write(content)

def get_repo_url():
	try:
		# Run the git command to get the remote URL
		result = subprocess.run(['git', 'config', '--get', 'remote.origin.url'],
		                        stdout=subprocess.PIPE,
		                        stderr=subprocess.PIPE,
		                        text=True,
		                        check=True)
		return result.stdout.strip()
	except subprocess.CalledProcessError as e:
		print(f"Error retrieving repository URL: {e.stderr}")
		return None

def get_module_name(repo_url: str) -> str:
	return repo_url.removeprefix("https://").removeprefix("http://").removeprefix("git@").removesuffix(".git")

def set_tool_name():
	repo_url = get_repo_url()
	TEMPLATE_VALUE = "<tool-name>"
	FILES = [
		"Dockerfile",
		".github/workflows/release.yml",
		"main.go",
		"cmd/version/command.go"
	]

	def is_tool_name_valid(value: str) -> bool:
		return re.match(r"^[a-zA-Z0-9_-]+$", value) is not None

	default_name = ""
	if repo_url:
		module_name = get_module_name(repo_url)
		default_name = module_name.split("/")[-1]


	while True:
		tool_name = str_prompt("Enter the tool name (alphanumeric, underscores, hyphens)", default_value=default_name)
		if is_tool_name_valid(tool_name):
			break
		print("Invalid tool name. Please use only alphanumeric characters, underscores, or hyphens.")

	replace_template(FILES, TEMPLATE_VALUE, tool_name)
	replace_template(["Makefile"], "OUTPUT_NAME := tool-name", f"OUTPUT_NAME := {tool_name}")

def set_module_name():
	repo_url = get_repo_url()
	TEMPLATE_VALUE = "github.com/username/template"
	FILES = [
		"go.mod",
		"main.go"
	]

	def is_valid_go_module_name(value: str) -> bool:
		return re.match(r"^[a-zA-Z0-9._/-]+$", value) is not None

	default_name = ""
	if repo_url:
		default_name = get_module_name(repo_url)


	while True:
		module_name = str_prompt("Enter the Go module name (alphanumeric, underscores, hyphens, slashes, dot)", default_name)
		if is_valid_go_module_name(module_name):
			break
		print("Invalid Go module name.")

	replace_template(FILES, TEMPLATE_VALUE, module_name)

def set_author():
	TEMPLATE_VALUE = "<author>"
	FILES = [
		"LICENSE"
	]

	# Get the current user's name
	try:
		user_name = subprocess.check_output(['git', 'config', 'user.name'], stderr=subprocess.STDOUT).decode().strip()
	except subprocess.CalledProcessError:
		user_name = ""

	author = str_prompt("Enter the author name", default_value=user_name)
	replace_template(FILES, TEMPLATE_VALUE, author)



def main():
	set_tool_name()
	set_module_name()
	set_author()

	current_year = datetime.datetime.now().year
	replace_template(["LICENSE"], "<year>", str(current_year))

	should_self_destruct = str_prompt("Do you want to delete this script? (y/n)", default_value="y")
	if should_self_destruct == "y":
		os.remove(__file__)




if __name__ == "__main__":
	main()
