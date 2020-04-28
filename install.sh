#!/bin/sh

main() {
	OS=$(detect_os)
	GOARCH=$(detect_goarch)
	GOOS=$(detect_goos)

	export shaper_INSTALLER=1

	SHAPER_BIN=$(bin_location)
	LATEST_RELEASE=$(get_release)

	while true; do
		CURRENT_RELEASE=$(get_current_release)
		log_debug "Start install loop with CURRENT_RELEASE=$CURRENT_RELEASE"

		if [ "$CURRENT_RELEASE" ]; then
			if [ "$CURRENT_RELEASE" != "$LATEST_RELEASE" ]; then
				log_debug "shaper is out of date ($CURRENT_RELEASE != $LATEST_RELEASE)"
				menu \
					u "Upgrade shaper from $CURRENT_RELEASE to $LATEST_RELEASE" upgrade \
					r "Remove shaper" uninstall \
					e "Exit" exit
			else
				log_debug "shaper is up to date ($CURRENT_RELEASE)"
				menu \
					r "Remove shaper" uninstall \
					e "Exit" exit
			fi
		else
			log_debug "shaper is not installed"
			menu \
				i "Install shaper" install \
				e "Exit" exit
		fi
	done
}

install() {
	log_info "Installing shaper..."
	if "install_bin"; then
		if [ ! -x "$SHAPER_BIN" ]; then
			log_error "Installation failed: binary not installed in $SHAPER_BIN"
			return 1
		fi
	fi
}

upgrade() {
	log_info "Upgrading shaper..."
	upgrade_bin
}

uninstall() {
	log_info "Uninstalling shaper..."
	uninstall_bin
}

install_bin() {
	bin_path=$SHAPER_BIN
	if [ "$1" ]; then
		bin_path=$1
	fi
	log_debug "Installing $LATEST_RELEASE binary for $GOOS/$GOARCH to $bin_path"
	url="https://github.com/ciokan/shaper/releases/download/v${LATEST_RELEASE}/shaper_${LATEST_RELEASE}_${GOOS}_${GOARCH}.tar.gz"
	log_debug "Downloading $url"
	mkdir -p "$(dirname "$bin_path")" &&
		curl -sL "$url" | asroot sh -c "tar Ozxf - shaper > \"$bin_path\"" &&
		asroot chmod 755 "$bin_path"
}

upgrade_bin() {
	tmp=$SHAPER_BIN.tmp
	if install_bin "$tmp"; then
		asroot mv -f "$tmp" "$SHAPER_BIN"
	fi
	rm -rf "$tmp"
}

uninstall_bin() {
	asroot "$SHAPER_BIN" uninstall
	asroot rm -f "$SHAPER_BIN"
}

log_debug() {
	if [ "$DEBUG" = "1" ]; then
		printf "\033[30;1mDEBUG: %s\033[0m\n" "$*" >&2
	fi
}

log_info() {
	printf "INFO: %s\n" "$*" >&2
}

log_error() {
	printf "\033[31mERROR: %s\033[0m\n" "$*" >&2
}

print() {
	# shellcheck disable=SC2059
	printf "$@" >&2
}

doc() {
	# shellcheck disable=SC2059
	printf "\033[30;1m%s\033[0m\n" "$*" >&2
}

menu() {
	while true; do
		n=0
		default=
		for item in "$@"; do
			case $((n % 3)) in
			0)
				key=$item
				if [ -z "$default" ]; then
					default=$key
				fi
				;;
			1)
				echo "$key) $item"
				;;
			esac
			n=$((n + 1))
		done
		print "Choice (default=%s): " "$default"
		read -r choice
		if [ -z "$choice" ]; then
			choice=$default
		fi
		n=0
		for item in "$@"; do
			case $((n % 3)) in
			0)
				key=$item
				;;
			2)
				if [ "$key" = "$choice" ]; then
					if ! "$item"; then
						log_error "$item: exit $?"
					fi
					break 2
				fi
				;;
			esac
			n=$((n + 1))
		done
		echo "Invalid choice"
	done
}

ask_bool() {
	msg=$1
	default=$2
	case $default in
	true)
		msg="$msg [Y|n]: "
		;;
	false)
		msg="$msg [y|N]: "
		;;
	*)
		msg="$msg (y/n): "
		;;
	esac
	while true; do
		print "%s" "$msg"
		read -r answer
		if [ -z "$answer" ]; then
			answer=$default
		fi
		case $answer in
		y | Y | yes | YES | true)
			echo "true"
			return 0
			;;
		n | N | no | NO | false)
			echo "false"
			return 0
			;;
		*)
			echo "Invalid input, use yes or no"
			;;
		esac
	done
}

detect_goarch() {
	if [ "$FORCE_GOARCH" ]; then
		echo "$FORCE_GOARCH"
		return 0
	fi
	case $(uname -m) in
	x86_64 | amd64)
		echo "amd64"
		;;
	i386 | i686)
		echo "386"
		;;
	*)
		log_error "Unsupported GOARCH: $(uname -m)"
		return 1
		;;
	esac
}

detect_goos() {
	if [ "$FORCE_GOOS" ]; then
		echo "$FORCE_GOOS"
		return 0
	fi
	case $(uname -s) in
	Linux)
		echo "linux"
		;;
	*)
		log_error "Unsupported GOOS: $(uname -s)"
		return 1
		;;
	esac
}

detect_os() {
	if [ "$FORCE_OS" ]; then
		echo "$FORCE_OS"
		return 0
	fi
	case $(uname -s) in
	Linux)
		case $(uname -o) in
		GNU/Linux)
			dist=$(
				. /etc/os-release
				echo "$ID"
			)
			case $dist in
			debian | ubuntu | elementary | raspbian | centos | fedora | rhel | arch | manjaro | openwrt | clear-linux-os | linuxmint | solus | pop)
				echo "$dist"
				return 0
				;;
			esac
			;;
		esac
		;;
	*) ;;
	esac
	log_error "Unsupported OS: $(uname -s)"
	return 1
}

asroot() {
	if [ "$(command -v sudo 2>/dev/null)" ]; then
		sudo "$@"
	else
		echo "Root required"
		su -m root -c "$*"
	fi
}

bin_location() {
	echo "/usr/bin/shaper"
}

get_current_release() {
	if [ -x "$SHAPER_BIN" ]; then
		$SHAPER_BIN version | cut -d' ' -f 3
	fi
}

get_release() {
	if [ "$shaper_VERSION" ]; then
		echo "$shaper_VERSION"
	else
		curl="curl -s"
		if [ -z "$(command -v curl >/dev/null 2>&1)" ]; then
			curl="openssl_get"
		fi
		$curl "https://api.github.com/repos/ciokan/shaper/releases/latest" |
			grep '"tag_name":' |
			sed -E 's/.*"([^"]+)".*/\1/' |
			sed -e 's/^v//'
	fi
}

openssl_get() {
	host=${1#https://*} # https://dom.com/path -> dom.com/path
	path=/${host#*/}    # dom.com/path -> /path
	host=${host%$path}  # dom.com/path -> dom.com
	printf "GET %s HTTP/1.0\nHost: %s\nUser-Agent: curl\n\n" "$path" "$host" |
		openssl s_client -quiet -connect "$host:443" 2>/dev/null
}

main
