require 'spec_helper'

# インストール済みチェック
%w{java curl python pip}.each do |binary|
  describe command("which #{binary}") do
    let(:disable_sudo) { true }
    its(:exit_status) { should eq 0}
  end
end

# python用パッケージインストール済みチェック
%w{pyaml docopt}.each do |package|
  describe command("pip show #{package}") do
    let(:disable_sudo) { true }
    its(:exit_status) { should eq 0}
  end
end

