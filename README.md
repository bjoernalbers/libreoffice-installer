# libreoffice-installer

A macOS Package to install the latest stable version of
[LibreOffice](https://www.libreoffice.org) for "business deployments."

## Features

It will (re-)install LibreOffice if the app...

- is not installed at all
- is outdated
- has been installed from the Mac App Store (see below why)

## Why this package?

I need to manage LibreOffice on multiple Macs and keep it up to date.
For business deployments one would normally use the Mac App Store version and
install it automatically via a mobile device management (MDM).
Unfortunately, this is not possible for our use case due to
[Bug 153927](https://bugs.documentfoundation.org/show_bug.cgi?id=153927).
This package automates the installation / update of LibreOffice from the
official download page.

## Usage

Just download the package (`libreoffice-installer-VERSION.pkg`) of the
[latest release](https://github.com/bjoernalbers/libreoffice-installer/releases/latest)
and install it manually or (better) deploy automatically via your MDM.
