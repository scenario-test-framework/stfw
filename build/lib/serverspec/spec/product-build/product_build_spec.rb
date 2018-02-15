require 'spec_helper'

# インストール済みチェック
%w{md5sum}.each do |binary|
  describe command("which #{binary}") do
    let(:disable_sudo) { true }
    its(:exit_status) { should eq 0}
  end
end
