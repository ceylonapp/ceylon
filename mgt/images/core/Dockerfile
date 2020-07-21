FROM python:3.7.8-buster

# set a directory for the app
WORKDIR /usr/src/app

# copy all the files to the container
COPY . .

RUN ls .
# install dependencies
RUN pip install --no-cache-dir -r config/requirements.txt

# define the port number the container should expose
EXPOSE 5000

RUN ls .
#RUN chmod +x ./run.sh
# run the command
CMD ["/bin/sh","run.sh"]
