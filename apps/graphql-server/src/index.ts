import { ApolloServer } from '@apollo/server';
import { startStandaloneServer } from '@apollo/server/standalone';
import { typeDefs } from './schema';
import { resolvers } from './resolvers';
import { ComplexityAnalyzer } from './middleware/complexityAnalyzer';
import * as fs from 'fs';
import * as path from 'path';

interface Context {
  complexity?: number;
  startTime?: number;
}

// Global complexity analyzer instance
const complexityAnalyzer = new ComplexityAnalyzer();

// File path for sharing complexity data between servers
const COMPLEXITY_DATA_FILE = path.join(__dirname, '..', 'complexity-data.json');

// Store complexity data to file for the REST server to read
function saveComplexityData() {
  const stats = complexityAnalyzer.getStats();
  const data = {
    lastUpdated: new Date().toISOString(),
    stats: stats,
    cacheSize: stats.cacheSize || 0,
    cachedQueries: stats.cachedQueries || []
  };
  
  try {
    fs.writeFileSync(COMPLEXITY_DATA_FILE, JSON.stringify(data, null, 2));
  } catch (error) {
    console.error('Failed to save complexity data:', error);
  }
}

async function startServer() {
  const server = new ApolloServer<Context>({
    typeDefs,
    resolvers,
    plugins: [
      // Complexity analysis plugin for REAL user queries
      {
        async requestDidStart() {
          return {
            async didResolveOperation(requestContext) {
              const complexity = complexityAnalyzer.analyze(
                requestContext.request.query || '',
                requestContext.request.variables
              );
              
              // Store complexity in context
              requestContext.contextValue.complexity = complexity;
              
              console.log(`[${new Date().toISOString()}] 🔍 REAL USER QUERY - Complexity: ${complexity}`);
              
              // Save complexity data to file for REST server
              saveComplexityData();
            },
          };
        },
      },
    ],
  });

  const PORT = Number(process.env.GRAPHQL_PORT) || 4000;

  const { url } = await startStandaloneServer(server, {
    listen: { port: PORT },
    context: async ({ req, res }) => {
      return {
        startTime: Date.now(),
      };
    },
  });

  console.log(`🚀 GraphQL Server ready at ${url}`);
  console.log(`🎪 Apollo Studio available at ${url}`);
  console.log(`📊 REST endpoints running on port ${PORT + 1}`);
  console.log(`💡 Visit ${url} to access Apollo Studio`);
  console.log(`🔍 Try these test queries:`);
  console.log(`   Simple:  query { user(id: "1") { name email } }`);
  console.log(`   Complex: query { systemAnalytics { totalUsers topUsers { name } } }`);
  
  // Start the REST server on the next port
  startRestServer(PORT + 1);
  
  // Log complexity analyzer stats periodically
  setInterval(() => {
    const stats = complexityAnalyzer.getStats();
    if (stats.cacheSize > 0) {
      console.log(`📈 Real user queries cached: ${stats.cacheSize}`);
      saveComplexityData(); // Update the shared data file
    }
  }, 10000);
}

// Start REST server for complexity stats
function startRestServer(port: number) {
  const http = require('http');
  const url = require('url');

  const restServer = http.createServer((req: any, res: any) => {
    const parsedUrl = url.parse(req.url, true);
    const path = parsedUrl.pathname;

    // Set CORS headers
    res.setHeader('Access-Control-Allow-Origin', '*');
    res.setHeader('Access-Control-Allow-Methods', 'GET, OPTIONS');
    res.setHeader('Access-Control-Allow-Headers', 'Content-Type');
    res.setHeader('Content-Type', 'application/json');

    if (req.method === 'OPTIONS') {
      res.writeHead(200);
      res.end();
      return;
    }

    if (path === '/health') {
      res.writeHead(200);
      res.end(JSON.stringify({
        status: 'healthy',
        timestamp: new Date().toISOString(),
        service: 'complexity-rest-server',
        version: '1.0.0'
      }));
      return;
    }

    if (path === '/complexity-stats') {
      try {
        // Read complexity data from file
        if (fs.existsSync(COMPLEXITY_DATA_FILE)) {
          const data = JSON.parse(fs.readFileSync(COMPLEXITY_DATA_FILE, 'utf8'));
          res.writeHead(200);
          res.end(JSON.stringify({
            cacheSize: data.stats.cacheSize || 0,
            cachedQueries: data.stats.cachedQueries || [],
            lastUpdated: data.lastUpdated
          }));
        } else {
          // No data file yet (no user queries)
          res.writeHead(200);
          res.end(JSON.stringify({
            cacheSize: 0,
            cachedQueries: [],
            lastUpdated: new Date().toISOString()
          }));
        }
      } catch (error) {
        res.writeHead(500);
        res.end(JSON.stringify({ error: 'Failed to read complexity data' }));
      }
      return;
    }

    if (path === '/metrics') {
      try {
        const data = fs.existsSync(COMPLEXITY_DATA_FILE) 
          ? JSON.parse(fs.readFileSync(COMPLEXITY_DATA_FILE, 'utf8'))
          : { stats: { cacheSize: 0, cachedQueries: [] } };
          
        res.writeHead(200);
        res.end(JSON.stringify({
          service: 'graphql-complexity-server',
          uptime: process.uptime(),
          memoryUsage: process.memoryUsage(),
          complexityStats: data.stats,
          lastUpdated: data.lastUpdated || new Date().toISOString()
        }));
      } catch (error) {
        res.writeHead(500);
        res.end(JSON.stringify({ error: 'Failed to read metrics' }));
      }
      return;
    }

    // 404 for unknown paths
    res.writeHead(404);
    res.end(JSON.stringify({ error: 'Not found' }));
  });

  restServer.listen(port, () => {
    console.log(`📊 REST endpoints server started on port ${port}`);
    console.log(`📊 Health check: http://localhost:${port}/health`);
    console.log(`📈 Complexity stats: http://localhost:${port}/complexity-stats`);
    console.log(`📊 Metrics: http://localhost:${port}/metrics`);
  });
}

// Graceful shutdown
process.on('SIGTERM', () => {
  console.log('SIGTERM received, shutting down gracefully');
  // Clean up complexity data file
  try {
    if (fs.existsSync(COMPLEXITY_DATA_FILE)) {
      fs.unlinkSync(COMPLEXITY_DATA_FILE);
    }
  } catch (error) {
    console.error('Error cleaning up complexity data file:', error);
  }
  process.exit(0);
});

process.on('SIGINT', () => {
  console.log('SIGINT received, shutting down gracefully');
  // Clean up complexity data file
  try {
    if (fs.existsSync(COMPLEXITY_DATA_FILE)) {
      fs.unlinkSync(COMPLEXITY_DATA_FILE);
    }
  } catch (error) {
    console.error('Error cleaning up complexity data file:', error);
  }
  process.exit(0);
});

// Start the server
startServer().catch((error) => {
  console.error('Failed to start server:', error);
  process.exit(1);
});

// Export complexity analyzer for external access
export { complexityAnalyzer };