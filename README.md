# libreoffice-installer

An installer package for macOS that deploys the latest stable version of
[LibreOffice](https://www.libreoffice.org).

## Features

It will automatically download and install LibreOffice if it...

- is not installed at all
- is outdated
- has been installed from the Mac App Store (see below)

The package detects the correct architecture (Intel vs. Apple Silicon) for each Mac.
It also attempts to quit LibreOffice if it is running.

## Why this package?

This is for Mac admins who need to manage LibreOffice on multiple Macs but cannot use the Mac App Store.
I created it because I couldn’t use the Mac App Store version due to
[Bug 153927](https://bugs.documentfoundation.org/show_bug.cgi?id=153927).

## Usage

1. Download the latest release:
   [libreoffice-installer.pkg](https://github.com/bjoernalbers/libreoffice-installer/releases/latest/download/libreoffice-installer.pkg)
2. Install the package on your fleet of Macs, preferably via Mobile Device Management (MDM).
