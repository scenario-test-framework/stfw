require 'spec_helper'

# インストール済みチェック
%w{java curl python pip}.each do |binary|
  describe command("which #{binary}") do
    let(:disable_sudo) { true }
    its(:exit_status) { should eq 0}
  end
end

# バージョンチェック
describe command('java -version') do
  let(:disable_sudo) { true }
  its(:stderr) { should match /version "1\.[89]\./ }
end
describe command('python --version') do
  let(:disable_sudo) { true }
  its(:stderr) { should match /Python 2\.7\./ }
end


# python用パッケージインストール済みチェック
%w{pyaml docopt}.each do |package|
  describe command("pip show #{package}") do
    let(:disable_sudo) { true }
    its(:exit_status) { should eq 0}
  end
end

