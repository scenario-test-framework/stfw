require 'spec_helper'

dir_module_bin = ENV["DIR_MODULE_BIN"]

%w{yaml2json json2yaml}.each do |target|
  describe command("which #{target}") do
    let(:disable_sudo) { true }
    its(:exit_status) { should eq 0}
  end

  describe file("#{dir_module_bin}/#{target}") do
    it { should be_file }
    it { should be_executable.by('owner') }
  end
end
