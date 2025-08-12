FROM node:18-alpine

WORKDIR /app

# Copy package files
COPY package*.json tsconfig.json ./

# Install with force (Docker handles conflicts better)
RUN npm install --legacy-peer-deps --force

# Copy source
COPY src/ ./src/

# Build
RUN npm run build

EXPOSE 4000 4001
CMD ["npm", "start"]