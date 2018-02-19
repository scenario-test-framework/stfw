require 'spec_helper'

path_digdag = ENV["PATH_DIGDAG"]

# permission
describe file("#{path_digdag}") do
  it { should be_file }
  it { should be_executable.by('owner') }
end
