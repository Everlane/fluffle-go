require 'rubygems'
require 'fluffle'

server = Fluffle::Server.new url: 'amqp://localhost'

server.drain do |dispatcher|
  dispatcher.handle('foo') { 'bar' }
end

server.start
