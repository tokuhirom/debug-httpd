FROM python:3.11-slim

WORKDIR /app

# Copy the server script
COPY server.py .

# Make the script executable
RUN chmod +x server.py

# Expose the default port
EXPOSE 9876

# Run the server with default port (can be overridden by CMD)
ENTRYPOINT ["python", "server.py"]
CMD ["9876"]