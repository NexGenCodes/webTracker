
import { NextAuthOptions } from "next-auth";
import CredentialsProvider from "next-auth/providers/credentials";

export const authOptions: NextAuthOptions = {
    providers: [
        CredentialsProvider({
            name: "Credentials",
            credentials: {
                username: { label: "Username", type: "text", placeholder: "admin" },
                password: { label: "Password", type: "password" }
            },
            async authorize(credentials, req) {
                const adminUser = process.env.ADMIN_USERNAME;
                const adminPass = process.env.ADMIN_PASSWORD;
                const adminEmail = process.env.ADMIN_EMAIL;

                if (credentials?.username === adminUser && credentials?.password === adminPass) {
                    return {
                        id: "admin-user",
                        name: adminUser,
                        email: adminEmail
                    };
                }
                return null;
            }
        })
    ],
    callbacks: {
        async session({ session, token }) {
            return session
        },
    },
    pages: {
        signIn: '/auth/signin',
    },
    secret: process.env.NEXTAUTH_SECRET,
};
