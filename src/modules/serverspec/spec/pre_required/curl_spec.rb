require 'spec_helper'

target = "curl"

# PATH
describe command("which #{target}") do
  let(:disable_sudo) { true }
  its(:exit_status) { should eq 0}
end
