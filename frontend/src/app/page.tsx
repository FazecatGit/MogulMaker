import { redirect } from 'next/navigation';

/**
 * Home Page (/)
 * 
 * This page automatically redirects to /dashboard
 * Users never see this - it's just a passthrough
 */

export default function Home() {
  redirect('/dashboard');
}
