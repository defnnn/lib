include ~/Makefile.common

get:
	earthly +get
	rsync -ia provider.new/. provider/.
	rm -rf provider.new

publish:
	passenv TWINE_USERNAME TWINE_PASSWORD -- p publish provider::
	passenv TWINE_USERNAME TWINE_PASSWORD -- p publish src:defn

package: dist/defn-$(shell cat src/defn/VERSION).tar.gz
	pipx install --force $<

dist/defn-$(shell cat src/defn/VERSION).tar.gz:
	echo >> src/BUILD
	./pants update-build-files src/BUILD

login-site:
	aws --profile gyre-ops sts get-caller-identity
	aws --profile curl-lib sts get-caller-identity
	aws --profile coil-lib sts get-caller-identity
	aws --profile helix-lib sts get-caller-identity
	aws --profile spiral-lib sts get-caller-identity

config-test:
	grep ^.profile ~/.aws/config | perl -ne 's{.profile }{}; s{.$$}{}; print if m{^[a-z]+-org$$}' | runmany 1 'echo profile "$$1 $$(aws --profile $$1 sts get-caller-identity | jq -r .Arn)"'
	@echo
	grep ^.profile ~/.aws/config | perl -ne 's{.profile }{}; s{.$$}{}; print if m{^\w+-\w+$$|-adm$$}' | runmany 1 'echo profile "$$1 $$(aws --profile $$1 sts get-caller-identity | jq -r .Arn)"'

config:
	earthly --push +config --stack=gyre   --region=us-east-2 --sso_region=us-east-2 --sso_url=https://d-9a6716e54a.awsapps.com/start
	earthly --push +config --stack=curl   --region=us-west-1 --sso_region=us-west-2 --sso_url=https://d-926760a859.awsapps.com/start
	earthly --push +config --stack=coil   --region=us-east-1 --sso_region=us-east-1 --sso_url=https://d-90674c3cfd.awsapps.com/start
	earthly --push +config --stack=helix  --region=us-east-2 --sso_region=us-east-2 --sso_url=https://d-9a6716ffd1.awsapps.com/start
	earthly --push +config --stack=spiral --region=us-west-2 --sso_region=us-west-2 --sso_url=https://d-926760b322.awsapps.com/start
	ls -d cdktf.*/stacks/*/.aws/config | sort | xargs cat > ~/.aws/config

plan-all:
	$(MAKE) plan stack=gyre
	$(MAKE) plan stack=curl
	$(MAKE) plan stack=coil
	$(MAKE) plan stack=helix
	$(MAKE) plan stack=spiral

synth:
	earthly +synth

plan:
	earthly +plan --stack=$(stack)

show:
	earthly +show --stack=$(stack)

import:
	earthly --push +import --stack=$(stack)

apply:
	earthly --push +apply --stack=$(stack)

buf:
	buf generate --include-wkt --include-imports
	cd p/defn/dev && $(MAKE)

pants-venv:
	cd 3rdparty/python && make all

pants-lock:
	-python -mvenv dist/export/python/virtualenvs/defn/3.10.6
	cd 3rdparty/python && make clean all

server client oper:
	./dist/src.defn/bean-$@.pex