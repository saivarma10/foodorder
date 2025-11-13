# Use official Python runtime as base image
FROM python:3.11-slim

# Set working directory
WORKDIR /app

# Copy requirements file (we'll create this)
COPY requirements.txt .

# Install dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY delivery.py .

# Expose port
EXPOSE 8010

# Run the application
CMD ["python", "delivery.py"]
