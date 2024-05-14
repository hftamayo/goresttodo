# Use the official PostgreSQL 14.x image as the base image
FROM postgres:14-alpine

# Set the working directory inside the container
WORKDIR /docker-entrypoint-initdb.d

# Copy the datalayerschema.sql script to the container
COPY datalayerschema.sql .

# Expose the PostgreSQL default port
EXPOSE 5432

# Run the script during container startup
CMD ["docker-entrypoint.sh", "postgres"]

#to run this file:
#docker build -t your-image-name .
#docker run --env-file .env your-image-name
#docker build -t hftamayo/pgrestapi:0.0.1 .
#docker run --name pgrestapi -d --env-file .env hftamayo/pgrestapi:0.0.1