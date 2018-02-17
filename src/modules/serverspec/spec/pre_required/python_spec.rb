require 'spec_helper'

target = "python"

# PATH
describe command("which #{target}") do
  let(:disable_sudo) { true }
  its(:exit_status) { should eq 0}
end

# version
describe command('python --version') do
  let(:disable_sudo) { true }
  its(:stderr) { should match /Python 2\.7\./ }
end


target = "pip"

# PATH
describe command("which #{target}") do
  let(:disable_sudo) { true }
  its(:exit_status) { should eq 0}
end

# pip package
%w{pyaml docopt}.each do |package|
  describe command("pip show #{package}") do
    let(:disable_sudo) { true }
    its(:exit_status) { should eq 0}
  end
end
