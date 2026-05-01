// Generate ES256 (P-256 ECDSA) key pair for Supabase JWT integration
import { generateKeyPairSync, createPublicKey } from 'crypto';
import { writeFileSync } from 'fs';
import { join } from 'path';

const { privateKey, publicKey } = generateKeyPairSync('ec', {
  namedCurve: 'P-256',
});

const privatePem = privateKey.export({ type: 'pkcs8', format: 'pem' });
const publicPem = publicKey.export({ type: 'spki', format: 'pem' });

// Save to backend directory
const backendDir = join(process.cwd(), '..', 'backend');
writeFileSync(join(backendDir, 'jwt_private.pem'), privatePem);
writeFileSync(join(backendDir, 'jwt_public.pem'), publicPem);

console.log('✅ Keys generated:');
console.log(`   Private: ${backendDir}/jwt_private.pem`);
console.log(`   Public:  ${backendDir}/jwt_public.pem`);
console.log('\n--- Private Key (for Supabase import) ---');
console.log(privatePem);
console.log('--- Public Key ---');
console.log(publicPem);

// Also generate JWK format for reference
const jwk = createPublicKey(publicPem).export({ format: 'jwk' });
console.log('--- Public Key (JWK format) ---');
console.log(JSON.stringify(jwk, null, 2));
