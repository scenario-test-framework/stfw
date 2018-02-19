require 'spec_helper'

target = "java"

# PATH
describe command("which #{target}") do
  let(:disable_sudo) { true }
  its(:exit_status) { should eq 0}
end

# version
describe command('java -version') do
  let(:disable_sudo) { true }
  its(:stderr) { should match /version "1\.[89]\./ }
end
