# Nutritional Value Scanner

A system for scanning and processing nutritional value labels using your phone and AI.
Once data is saved, you can keep track of how many calories, fats etc are netering your home. 

## Requirements

- Go 1.21 or higher
- 
### For local ML processing
- NVIDIA GPU drivers
- CUDA toolkit

## Setup

After cloning this repository

### AI setup

#### Google

1. Ensure you have an AI-enabled project
1. Ensure you have access to a credentials file ([see how to create it](https://developers.google.com/workspace/guides/create-credentials))
1. Create in `backend/config/`a `google.json` with the fields 
```json
{
    "project_id": <your project name>,
    "location": <your project location>,
    "credentials_file": <your credentials file path>
} 
```

### Running:
1. Open terminal and move to the `backend`folder
1. Run the server using `go run cmd/server/main.go`
1. If asked, allow the server to access your network
1. Access the web interface at `http://<server IP>:<port set in config.json>`

## Architecture

The system uses a client-server architecture where:
- Webpage captures images and sends them to the desktop
- Desktop server processes images using ML models
- Results are stored in a local database
- Real-time updates are sent back to the mobile app

## License

MIT License