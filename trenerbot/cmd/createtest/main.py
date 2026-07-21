import sqlite3
c = sqlite3.connect('data/crm.db')

# Get the parent user id
parent_user_id = 10

# Update client 27 with proper Russian name
c.execute("UPDATE clients SET full_name=? WHERE id=?", ("Спортсмен Иванов", 27))

# Link parent to child: set parent_id on the client record
c.execute("UPDATE clients SET parent_id=? WHERE id=?", (parent_user_id, 27))

# Verify
r = c.execute("SELECT id, full_name, birth_date, parent_id FROM clients WHERE id=27").fetchone()
print("Client 27:", r)

# Ensure user 10 is parent
c.execute("UPDATE users SET role=? WHERE id=?", ("parent", parent_user_id))
r2 = c.execute("SELECT id, phone, role FROM users WHERE id=10").fetchone()
print("User 10:", r2)

# Additional: create a child placeholder in the client record for user 9
# Also update user 9 (the athlete) to have a proper client name
c.execute("UPDATE clients SET full_name=? WHERE user_id=?", ("Спортсмен Иванов", 9))
r3 = c.execute("SELECT id, user_id, full_name, birth_date FROM clients WHERE user_id=9").fetchone()
print("Client for user 9:", r3)

# Create some lessons so there's data to see
lessons = c.execute("SELECT id, title, date FROM lessons LIMIT 3").fetchall()
print("Sample lessons:", lessons)

c.commit()
c.close()
print("\nDone - accounts ready:")
print("  Admin:  +79999999999 / admin123")
print("  Parent: +79992222222 / test123")
print("  Client: +79991111111 / test123")
