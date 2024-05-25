# Use an official Python runtime as a parent image
FROM python:3.9-slim

# Set the working directory in the container
WORKDIR /usr/src/app

# Copy the current directory contents into the container at /usr/src/app
COPY . .

# Install any needed packages specified in requirements.txt
# Since your script does not require external packages, this step is optional
# If needed, uncomment the following lines and create a requirements.txt file
# COPY requirements.txt ./
# RUN pip install --no-cache-dir -r requirements.txt

# Make port 9999 available to the world outside this container
EXPOSE 9999

# Run 5G_traffic_detection.py when the container launches
CMD ["python", "./5G_traffic_detection.py"]