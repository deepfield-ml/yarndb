import random
import yaml
import os
from faker import Faker
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Any, Optional, Callable

# Initialize Faker for generating fake data
fake = Faker()

class YarnDBDataGenerator:
    """
    Generates a complete, consistent set of fake data for a YarnDB-style database.

    This class creates various types of records (users, products, orders, etc.)
    and ensures that relationships between them (e.g., an order's user_id)
    are valid by using previously generated record IDs.

    The output can be saved as individual YAML files and/or a shell script
    with `yarndb` CLI commands to populate a database.
    """

    def __init__(self, output_dir: str = "yarndb_generated_data"):
        """
        Initializes the data generator.

        Args:
            output_dir (str): The directory where generated files will be saved.
                              This directory will be created if it doesn't exist.
        """
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(parents=True, exist_ok=True)
        
        self.record_types: Dict[str, Callable[..., Dict[str, Any]]] = {
            'users': self._generate_user,
            'products': self._generate_product,
            'orders': self._generate_order,
            'categories': self._generate_category,
            'employees': self._generate_employee
        }
        
        self.category_name_pool = list(set([fake.bs().replace(' ', '_') for _ in range(200)]))
        random.shuffle(self.category_name_pool)

    def _generate_user(self, record_id: str) -> Dict[str, Any]:
        """Generates a single user record."""
        return {
            'name': fake.name(),
            'email': fake.unique.email(),
            'department': random.choice(['engineering', 'marketing', 'sales', 'hr']),
            'age': random.randint(22, 65),
            'skills': random.sample(['Python', 'Go', 'Docker', 'Kubernetes', 'React', 'SQL', 'Terraform'],
                                  random.randint(2, 4)),
            'created_at': fake.date_time_between(start_date='-2y', end_date='now').isoformat()
        }

    def _generate_product(self, record_id: str, category_ids: List[str]) -> Dict[str, Any]:
        """Generates a single product record, linked to an existing category."""
        return {
            'name': fake.catch_phrase(),
            'price': round(random.uniform(10.0, 999.99), 2),
            'category_id': random.choice(category_ids) if category_ids else None,
            'description': fake.text(max_nb_chars=200),
            'in_stock': random.choice([True, False]),
            'attributes': {
                'weight': f"{random.uniform(0.1, 10.0):.1f}kg",
                'color': fake.color_name(),
                'material': random.choice(['plastic', 'metal', 'wood', 'fabric', 'ceramic'])
            }
        }

    def _generate_order(self, record_id: str, user_ids: List[str], product_ids: List[str]) -> Dict[str, Any]:
        """Generates a single order record, linked to existing users and products."""
        return {
            'user_id': random.choice(user_ids) if user_ids else None,
            'product_ids': random.sample(product_ids, k=random.randint(1, min(5, len(product_ids)))) if product_ids else [],
            'total_amount': round(random.uniform(25.0, 500.0), 2),
            'status': random.choice(['pending', 'processing', 'shipped', 'delivered', 'cancelled']),
            'order_date': fake.date_time_between(start_date='-1y', end_date='now').isoformat(),
            'shipping_address': {
                'street': fake.street_address(),
                'city': fake.city(),
                'state': fake.state_abbr(),
                'zip_code': fake.zipcode()
            }
        }

    def _generate_category(self, record_id: str, existing_category_ids: List[str]) -> Dict[str, Any]:
        """Generates a single category record, with an optional parent category."""
        name = self.category_name_pool.pop() if self.category_name_pool else f"fallback_category_{random.randint(1000, 9999)}"
        
        has_parent = random.choice([True, False]) and existing_category_ids
        parent_category = random.choice(existing_category_ids) if has_parent else None
        
        return {
            'name': name,
            'description': fake.text(max_nb_chars=150),
            'parent_category': parent_category
        }

    def _generate_employee(self, record_id: str, existing_employee_ids: List[str]) -> Dict[str, Any]:
        """Generates a single employee, with an optional manager from existing employees."""
        possible_managers = [eid for eid in existing_employee_ids if eid != record_id]
        has_manager = random.choice([True, False]) and possible_managers
        manager_id = random.choice(possible_managers) if has_manager else None

        return {
            'name': fake.name(),
            'employee_id': fake.uuid4(),
            'department': random.choice(['engineering', 'marketing', 'sales', 'hr', 'finance', 'operations']),
            'position': fake.job(),
            'salary': random.randint(40000, 150000),
            'hire_date': fake.date_between(start_date='-5y', end_date='now').isoformat(),
            'manager_id': manager_id
        }
    
    # --- THIS IS THE CORRECTED METHOD ---
    def generate_records(self, record_type: str, count: int, **kwargs: Any) -> Dict[str, Dict[str, Any]]:
        """
        Generates a specified number of records for a given type.
        """
        if record_type not in self.record_types:
            raise ValueError(f"Unknown record type: {record_type}")

        records: Dict[str, Dict[str, Any]] = {}
        generator_func = self.record_types[record_type]
        
        # Special handling for types that can reference themselves (e.g., manager_id)
        # We build up a list of IDs as we generate them and pass it to the generator.
        if record_type == 'categories':
            temp_id_list = []
            for i in range(1, count + 1):
                record_id = f"{record_type}_{i}"
                # Call with the correct keyword argument: 'existing_category_ids'
                records[record_id] = generator_func(record_id, existing_category_ids=temp_id_list)
                temp_id_list.append(record_id)
            return records

        if record_type == 'employees':
            temp_id_list = []
            for i in range(1, count + 1):
                record_id = f"{record_type}_{i}"
                # Call with the correct keyword argument: 'existing_employee_ids'
                records[record_id] = generator_func(record_id, existing_employee_ids=temp_id_list)
                temp_id_list.append(record_id)
            return records

        # Default handling for all other record types
        for i in range(1, count + 1):
            record_id = f"{record_type}_{i}"
            records[record_id] = generator_func(record_id, **kwargs)

        return records

    def save_to_yaml_files(self, database: Dict[str, Dict[str, Any]]):
        """Saves generated records to separate YAML files for each record type."""
        for record_type, records in database.items():
            filepath = self.output_dir / f"records_{record_type}.yaml"
            with open(filepath, 'w') as f:
                yaml.dump(records, f, default_flow_style=False, indent=2, sort_keys=False)
            print(f"✅ Generated {len(records)} {record_type} records in {filepath}")

    def generate_database(self, config: Optional[Dict[str, int]] = None) -> Dict[str, Dict[str, Any]]:
        """Generates a complete database with consistent, related records."""
        if config is None:
            config = {
                'categories': 10,
                'users': 100,
                'employees': 30,
                'products': 50,
                'orders': 200,
            }
        
        print(f"--- Starting Database Generation in '{self.output_dir}' ---")
        database: Dict[str, Dict[str, Any]] = {}
        
        generation_order = ['categories', 'users', 'employees', 'products', 'orders']
        id_map: Dict[str, List[str]] = {}

        for record_type in generation_order:
            if record_type in config:
                count = config[record_type]
                print(f"⚙️  Generating {count} {record_type} records...")
                
                dependencies = {}
                if record_type == 'products':
                    dependencies['category_ids'] = id_map.get('categories_ids', [])
                elif record_type == 'orders':
                    dependencies['user_ids'] = id_map.get('users_ids', [])
                    dependencies['product_ids'] = id_map.get('products_ids', [])

                records = self.generate_records(record_type, count, **dependencies)
                database[record_type] = records
                id_map[f"{record_type}_ids"] = list(records.keys())
        
        self.save_to_yaml_files(database)
        print("--- Database Generation Complete ---")
        return database

    def generate_yarndb_commands(self, database: Dict[str, Dict[str, Any]], output_file: str = "populate_yarndb.sh"):
        """Generates a shell script with YarnDB CLI commands to populate the database."""
        filepath = Path(output_file)
        commands = [
            "#!/bin/bash",
            "# YarnDB database population script",
            "# Generated by YarnDBDataGenerator",
            "set -e # Exit immediately if a command exits with a non-zero status.",
            "",
            "echo 'Initializing YarnDB...'",
            "yarndb init",
            "",
            "echo 'Creating indexes...'",
            "yarndb index department",
            "yarndb index category_id",
            "yarndb index status",
            "yarndb index user_id",
            "",
            "echo 'Inserting records...'",
        ]

        total_records = sum(len(records) for records in database.values())
        for record_type, records in database.items():
            commands.append(f"\n# Inserting {record_type} records...")
            for record_id, data in records.items():
                yaml_data = yaml.dump(data, default_flow_style=False, sort_keys=False)
                escaped_yaml = yaml_data.replace('\\', '\\\\').replace("'", "\\'")
                commands.append(f'yarndb set {record_id} $\'{escaped_yaml}\'')
                
        commands.extend([
            "",
            "echo 'Saving database to disk...'",
            "yarndb save",
            "",
            f"echo '✅ Successfully populated YarnDB with {total_records} records.'"
        ])

        with open(filepath, 'w') as f:
            f.write('\n'.join(commands))
        
        os.chmod(filepath, 0o755)

        print(f"✅ Generated YarnDB commands in executable script: ./{filepath}")

def main():
    """Main function to run the data generation process."""
    generator = YarnDBDataGenerator(output_dir="yarndb_generated_data")

    db_config = {
        'categories': 8,
        'users': 50,
        'employees': 25,
        'products': 40,
        'orders': 75,
    }

    database_records = generator.generate_database(db_config)
    generator.generate_yarndb_commands(database_records)

    print("\nGeneration complete!")
    print("Next steps:")
    print("1. Review the generated files in the 'yarndb_generated_data/' directory.")
    print("2. Run the shell script to populate your YarnDB instance: ./populate_yarndb.sh")

if __name__ == "__main__":
    main()