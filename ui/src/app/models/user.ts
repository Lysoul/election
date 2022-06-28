export interface User {
    national_id: string;
    full_name: string;
    email: string;
    password_changed_at?: string;
    create_at?: string;
  }


export const DefaultUser: User = {
    national_id: '-1',
    full_name: 'Anonymouse',
    email: '',
}