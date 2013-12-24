package sdeb

const (
	TEMPLATE_DEBIAN_COMPAT         = "9"
	FORMAT_DEFAULT                 = "3.0 (quilt)"
	TEMPLATE_DEBIAN_SOURCE_FORMAT  = FORMAT_DEFAULT
	TEMPLATE_DEBIAN_SOURCE_OPTIONS = `tar-ignore = .hg
tar-ignore = .git
tar-ignore = .bzr`

	TEMPLATE_DEBIAN_RULES = `#!/usr/bin/make -f
# -*- makefile -*-

# Uncomment this to turn on verbose mode.
#export DH_VERBOSE=1

export GOPATH=$(CURDIR)

PKGDIR=debian/{{.PackageName}}

%:
	dh $@

clean:
	dh_clean
	rm -rf $(GOPATH)/bin/* $(GOPATH)/pkg/*
	#cd $(GOPATH)/src && find * -name '*.go' -exec dirname {} \; | xargs -n1 go clean
	rm -f $(GOPATH)/goinstall.log

binary-arch: clean
	dh_prep
	dh_installdirs
	cd $(GOPATH)/src && find * -name '*.go' -exec dirname {} \; | xargs -n1 go install
	mkdir -p $(PKGDIR)/usr/bin
	cp $(GOPATH)/bin/* $(PKGDIR)/usr/bin/
	dh_strip
	dh_compress
	dh_fixperms
	dh_installdeb
	dh_gencontrol
	dh_md5sums
	dh_builddeb

binary: binary-arch`

	TEMPLATE_SOURCEDEB_CONTROL = `Source: {{.PackageName}}
Build-Depends: {{.BuildDepends}}
Priority: {{.Priority}}
Maintainer: {{.Maintainer}}
Standards-Version: {{.StandardsVersion}}
Section: {{.Section}}

Package: {{.PackageName}}
Architecture: {{.Architecture}}
Depends: ${misc:Depends}{{.Depends}}
Description: {{.Description}}
{{.Other}}`

	TEMPLATE_DEBIAN_DSC = `Format: {{.Format}}
Source: {{.PackageName}}
Binary: {{.PackageName}}
Architecture: {{.Architecture}}
Version: {{.PackageVersion}}
Maintainer: {{.Maintainer}}
Standards-Version: {{.StandardsVersion}}
Build-Depends: {{.BuildDepends}}
Priority: {{.Priority}}
Section: {{.Section}}
Checksums-Sha1:{{range .ChecksumsSha1}}
 {{.Checksum}} {{.Size}} {{.File}}{{end}}
Checksums-Sha256:{{range .ChecksumsSha256}}
 {{.Checksum}} {{.Size}} {{.File}}{{end}}
Files:{{range .Files}}
 {{.Checksum}} {{.Size}} {{.File}}{{end}}
{{.Other}}`

	TEMPLATE_CHANGELOG_HEADER        = `{{.PackageName}} ({{.PackageVersion}}) {{.Status}}; urgency=low`
	TEMPLATE_CHANGELOG_INITIAL_ENTRY = `  * Initial import`
	TEMPLATE_CHANGELOG_FOOTER        = ` -- {{.Maintainer}} <{{.MaintainerEmail}}>  {{.EntryDate}}`
	TEMPLATE_DEBIAN_COPYRIGHT        = `Copyright 2013 {{.PackageName}}`
	TEMPLATE_DEBIAN_README           = `{{.PackageName}}
==========

`

	STATUS_DEFAULT            = "unreleased"
	SECTION_DEFAULT           = "devel" //TODO: correct to use this?
	PRIORITY_DEFAULT          = "extra"
	BUILD_DEPENDS_DEFAULT     = "debhelper (>= 9.1.0), golang-go"
	STANDARDS_VERSION_DEFAULT = "3.9.4"
	ARCHITECTURE_DEFAULT      = "any"
	DIRNAME_TEMP              = ".goxc-temp"
)
