package sdeb

const (
	FILETEMPLATE_DEBIAN_RULES = `#!/usr/bin/make -f
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

	SECTION_DEFAULT           = "devel" //TODO: correct to use this?
	PRIORITY_DEFAULT          = "extra"
	BUILD_DEPENDS_DEFAULT     = "debhelper (>= 7.0.50~), golang-stable"
	STANDARDS_VERSION_DEFAULT = "3.9.1"
	ARCHITECTURE_DEFAULT      = "amd64 i386 armel"
	DIRNAME_TEMP              = ".goxc-temp"
)
