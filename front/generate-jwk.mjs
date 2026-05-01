import { createPrivateKey } from 'crypto';
import { readFileSync } from 'fs';
import { join } from 'path';

// Read the PEM we already generated
const backendDir = join(process.cwd(), '..', 'backend');
const privatePem = readFileSync(join(backendDir, 'jwt_private.pem'));

// Convert to Private JWK
const privateJwk = createPrivateKey(privatePem).export({ format: 'jwk' });

// Supabase might require a "kid" (Key ID) property in the JWK to match our backend
privateJwk.kid = "3ac00c7e-2058-4c54-8cf1-54ebca7a67f1";
// Supabase might require the alg property too
privateJwk.alg = "ES256";

console.log('--- Copy this JSON into Supabase ---');
console.log(JSON.stringify(privateJwk, null, 2));
