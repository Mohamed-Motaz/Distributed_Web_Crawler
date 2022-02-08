# Distributed_Web_Crawler

If you are on MacOs
first install brew

run 
brew update
brew doctor
brew install postgresql



 brew services start postgres


to start rabbitmq, docker run --name rabbitmq-container -p 5672:5672 rabbitmq

I use exponential backoff when a worker asks for a respnose from master and doesn't get an appropriate one

Each worker keeps his finished job for only 10 seconds, and retry within those 10 seconds to keep sending the data once every second, then the worker destroys the data and attempts to ask for a new task
