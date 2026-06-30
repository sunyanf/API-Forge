// Node.js fetch example (node 18+ or use node-fetch)
const fetch = (...args) => import('node-fetch').then(({default: f}) => f(...args));

async function main(){
  const res = await fetch('http://localhost:8080/api/v1/me', {
    headers: { 'Authorization': 'Bearer YOUR_JWT_HERE' }
  });
  if(!res.ok){
    console.error('status', res.status);
    process.exit(1);
  }
  const data = await res.json();
  console.log('profile', data);
}

main().catch(console.error);
