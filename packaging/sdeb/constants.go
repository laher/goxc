package sdeb

const (
	TEMPLATE_DEBIAN_RULES = `#!/usr/bin/make -f
# -*- makefile -*-

# Uncomment this to turn on verbose mode.
#export DH_VERBOSE=1

export GOPATH=$(CURDIR)

PKGDIR=debian/{{.}}

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
	FILECONTENT_DEBIAN_COMPAT         = "7"            //TODO: grok significance
	FILECONTENT_DEBIAN_SOURCE_FORMAT  = "3.0 (native)" //TODO: grok significance
	FILECONTENT_DEBIAN_SOURCE_OPTIONS = `tar-ignore = .hg
tar-ignore = .git
tar-ignore = .bzr`
	TEMPLATE_SOURCEDEB_CONTROL =`Source: {{.PackageName}}
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

	TEMPLATE_CHANGELOG_HEADER = `{{.PackageName}} ({{.StandardsVersion}} {{.Status}}; urgency=low`
	TEMPLATE_CHANGELOG_INITIAL_ENTRY = ` * Initial import
`
	TEMPLATE_CHANGELOG_FOOTER = ` -- {{.Maintainer}} <{{.MaintainerEmail}}>  {{.EntryDate}}`
	TEMPLATE_STATUS_DEFAULT   = "unreleased"

	SECTION_DEFAULT           = "devel" //TODO: correct to use this?
	PRIORITY_DEFAULT          = "extra"
	BUILD_DEPENDS_DEFAULT     = "debhelper (>= 9.1.0), golang-stable"
	STANDARDS_VERSION_DEFAULT = "3.9.4"
	ARCHITECTURE_DEFAULT      = "any"
	DIRNAME_TEMP              = ".goxc-temp"
)
