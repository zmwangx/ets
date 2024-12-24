class Ets < Formula
  desc "Command output timestamper"
  homepage "https://github.com/zmwangx/ets"
  url "https://github.com/zmwangx/ets/archive/v0.2.2.tar.gz"
  sha256 "e1c0575c1b96ecf34cbd047efacf78b57a03fc1d4ac805f5f59e4e75e51f78d0"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args, "-ldflags", "-X main.version=#{version}"
  end

  test do
    assert_match "[00:00:00]", shell_output("#{bin}/ets -s echo hello").chomp
  end
end
